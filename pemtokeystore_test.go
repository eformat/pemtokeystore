//  Copyright 2016 Red Hat, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package pemtokeystore_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jimmidyson/pemtokeystore"
)

const (
	rootCAFile                   = "root-ca"
	intermediateCAFile           = "intermediate-ca"
	serverFromRootCAFile         = "server-from-root"
	serverFromIntermediateCAFile = "server-from-intermediate"
)

var testKeystore = filepath.Join("testdata", "test.ks")

func certFile(name string) string {
	return filepath.Join("testdata", name+".pem")
}

func keyFile(name string) string {
	return filepath.Join("testdata", name+"-key.pem")
}

func startTLSServer(certs ...string) (*httptest.Server, error) {
	var certChainBytes []byte
	for _, cert := range certs {
		b, err := ioutil.ReadFile(certFile(cert))
		if err != nil {
			return nil, err
		}
		certChainBytes = append(certChainBytes, b...)
	}
	keyBytes, err := ioutil.ReadFile(keyFile(certs[0]))
	if err != nil {
		return nil, err
	}
	serverCert, err := tls.X509KeyPair(certChainBytes, keyBytes)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))

	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}
	ts.StartTLS()

	// Validate server is correctly set up using golang client and only specifiying root cert as trusted.
	caCert, err := ioutil.ReadFile(certFile(certs[len(certs)-1]))
	if err != nil {
		ts.Close()
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}
	resp, err := client.Get(ts.URL)
	if err != nil {
		ts.Close()
		return nil, err
	}
	resp.Body.Close()

	return ts, nil
}

func TestJavaValidateServerFromRootCA(t *testing.T) {
	serverCerts := [][]string{
		[]string{serverFromRootCAFile, rootCAFile},
		[]string{serverFromIntermediateCAFile, intermediateCAFile, rootCAFile},
	}

	javac, err := exec.LookPath("javac")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(javac, "testdata/Client.java")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Fatal(err)
	}

	java, err := exec.LookPath("java")
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range serverCerts {
		func() {
			ts, err := startTLSServer(s...)
			if err != nil {
				t.Error(err)
				return
			}
			defer ts.Close()

			opts := pemtokeystore.Options{
				CACertFiles:  []string{certFile(rootCAFile)},
				KeystorePath: testKeystore,
			}
			//defer os.Remove(testKeystore)
			if err = pemtokeystore.CreateKeystore(opts); err != nil {
				t.Error(err)
				return
			}
			if err = validateKeystoreWithKeytool(t); err != nil {
				t.Error(err)
				return
			}

			cmd := exec.Command(java, "-cp", "testdata", "Client", testKeystore, ts.URL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Log(string(out))
				t.Error(err)
				return
			}
		}()
	}
}

func validateKeystoreWithKeytool(t *testing.T) error {
	keytool, err := exec.LookPath("keytool")
	if err == nil {
		cmd := exec.Command(keytool, "-list", "-keystore", testKeystore, "-storepass", "")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		return err
	}
	return nil
}
