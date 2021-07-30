/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testsuites

import (
	"context"
	"strings"

	"github.com/onsi/ginkgo"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	storageframework "k8s.io/kubernetes/test/e2e/storage/framework"
	storageutils "k8s.io/kubernetes/test/e2e/storage/utils"
)

type authTestSuite struct {
	tsInfo storageframework.TestSuiteInfo
}

// InitauthTestSuite returns authTestSuite that implements TestSuite interface
func InitCustomAuthTestSuite(patterns []storageframework.TestPattern) storageframework.TestSuite {
	return &authTestSuite{
		tsInfo: storageframework.TestSuiteInfo{
			Name:         "auth",
			TestPatterns: patterns,
		},
	}
}

func InitAuthTestSuite() storageframework.TestSuite {
	patterns := []storageframework.TestPattern{
		storageframework.AuthDynamicPV,
	}
	return InitCustomAuthTestSuite(patterns)
}

func (s *authTestSuite) GetTestSuiteInfo() storageframework.TestSuiteInfo {
	return s.tsInfo
}

func (s *authTestSuite) SkipUnsupportedTests(driver storageframework.TestDriver, pattern storageframework.TestPattern) {
	dInfo := driver.GetDriverInfo()
	_, ok := driver.(storageframework.DynamicPVTestDriver)
	if !ok {
		e2eskipper.Skipf("Driver %s doesn't support %v -- skipping", dInfo.Name, pattern.VolType)
	}

	_, ok = driver.(storageframework.AuthTestDriver)
	if !ok {
		e2eskipper.Skipf("Driver %s does not support auth -- skipping", dInfo.Name)
	}
}

func (a *authTestSuite) DefineTests(driver storageframework.TestDriver, pattern storageframework.TestPattern) {
	type local struct {
		driverInfo    *storageframework.DriverInfo
		config        *storageframework.PerTestConfig
		driverCleanup func()

		resource  *storageframework.VolumeResource
		podConfig *e2epod.Config

		scAuthParams   map[string]string
		authSecretData map[string]string
		authMatchGroup [][]storageframework.CSIStorageClassAuthParamKey
	}
	var l local

	// Beware that it also registers an AfterEach which renders f unusable. Any code using
	// f must run inside an It or Context callback.
	f := framework.NewFrameworkWithCustomTimeouts("auth", storageframework.GetDriverTimeouts(driver))

	init := func() {
		l = local{}

		l.driverInfo = driver.GetDriverInfo()

		authDriver, ok := driver.(storageframework.AuthTestDriver)
		framework.ExpectEqual(ok, true, "Driver not yet implement interface: AuthTestDriver")
		l.authSecretData = authDriver.GetAuthSecretData()
		l.authMatchGroup = authDriver.GetAuthMatchGroup()

		// Now do the more expensive test initialization.
		l.config, l.driverCleanup = driver.PrepareTest(f)
		testVolumeSizeRange := a.GetTestSuiteInfo().SupportedSizeRange

		l.scAuthParams = authDriver.GetStorageClassAuthParameters(l.config)

		// If authentication is required in the CSI Provisioner phase, create a
		// secret first for create pvc in CreateVolumeResource()
		secretName, ok := l.scAuthParams[string(storageframework.CSIProvisionerSecretName)]
		if ok {
			err := createSecret(f.ClientSet, f.Namespace.Name,
				secretName, l.authSecretData)
			framework.ExpectNoError(err, "Failed to create provisioner secret")
		}
		l.resource = storageframework.CreateVolumeResource(driver, l.config, pattern, testVolumeSizeRange)

		l.podConfig = &e2epod.Config{
			NS:            f.Namespace.Name,
			PVCs:          []*v1.PersistentVolumeClaim{l.resource.Pvc},
			SeLinuxLabel:  e2epod.GetLinuxLabel(),
			NodeSelection: l.config.ClientNodeSelection,
			ImageID:       e2epod.GetDefaultTestImageID(),
		}
	}

	cleanup := func() {
		var errs []error

		if l.resource != nil {
			errs = append(errs, l.resource.CleanupResource())
			l.resource = nil
		}

		errs = append(errs, storageutils.TryFunc(l.driverCleanup))
		framework.ExpectNoError(errors.NewAggregate(errs), "while cleaning up resource")
	}

	ginkgo.It("pod should attach pvc success using current authSecretData", func() {
		init()
		defer cleanup()
		for paramKey, paramValue := range l.scAuthParams {
			if strings.HasSuffix(paramKey, "secret-name") {
				secretName := paramValue
				err := createSecret(f.ClientSet, f.Namespace.Name,
					secretName, l.authSecretData)
				framework.ExpectNoError(err, "Failed to create secret: ", secretName)
			}
		}
		pod, err := e2epod.CreateSecPod(f.ClientSet, l.podConfig, f.Timeouts.PodStartShort)
		framework.ExpectNoError(err)
		defer func() {
			framework.ExpectNoError(e2epod.DeletePodWithWait(f.ClientSet, pod))
		}()
	})
}

func createSecret(kubeClient clientset.Interface, namespace, name string, data map[string]string) error {
	secret := makeSecret(namespace, name)
	secret.StringData = data

	if createdSecret, err := kubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{}); err == nil {
		framework.Logf("Updating secret %v in ns %v", name, namespace)
		createdSecret.Data = secret.Data
		_, err = kubeClient.CoreV1().Secrets(namespace).Update(context.TODO(), createdSecret, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	framework.Logf("Creating secret %v in ns %v", name, namespace)
	_, err := kubeClient.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func makeSecret(namespace, name string) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return secret
}
