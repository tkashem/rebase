# rebase


steps:
1. bump openshift/api, open a PR 
(it has dependency on k8s.io repos, no dependency on openshift/* repo)   

2. bump openshift/client-go, open a PR
- bump k8s.io version
- bump github.com/openshift/api (pin to the branch you created in step 1 temporarily)


3. bump openshift/library-go, open a PR
- bump k8s.io version
- bump github.com/openshift/api (pin to the branch you created in step 1 temporarily) 
- bump github.com/openshift/client-go (pin to the branch you created in step 2 temporarily)

4. bump openshift/apiserver-library-go, open a PR
- bump k8s.io version
- bump github.com/openshift/api (pin to the branch you created in step 1 temporarily)
- bump github.com/openshift/client-go (pin to the branch you created in step 2 temporarily)
- bump github.com/openshift/library-go (pin to the branch you created in step 3 temporarily)


5. bring the carry commits in the new o/k branch
6. add a new commit that updates the k8s version, golang version, and proper builder image (if applicable)
make sure you only specify the major.minor.patch of the target version. for example, if you are
using tag `v1.24.0-rc-0` then specify `1.24.0` as the kubernetes version in Dockerfile.rhel
```
LABEL io.openshift.build.versions="kubernetes=1.24.0"
```   

7. create a new commit with the following:
- A: hack/pin-dependency.sh openshift/api {your branch from 1}
     hack/update-vendor.sh
- B: hack/pin-dependency.sh openshift/client-go {your branch from 2}
     hack/update-vendor.sh 
- C: hack/pin-dependency.sh openshift/library-go {your branch from 3}
     hack/update-vendor.sh
- B: hack/pin-dependency.sh openshift/apiserver-library-go {your branch from 4}
     hack/update-vendor.sh


hack/update-vendor.sh
manually add replace for o/api, o/client-go ..

(commit the go.mod file)

go mod tidy
hack/update-vendor.sh

staging/code-genarator directory - go mod vendor



8. make build?
9. make update?


go 1.18.1? hack/libgolang.sh 484 minimum_go_version


protoc 3.0.0

10. make test

```
$ GITHUB_AUTH_TOKEN={your github access token} /home/akashem/go/src/github.com/tkashem/rebase/output/rebase apply --target=v1.24 --carry-commit-file=/home/akashem/go/src/github.com/tkashem/rebase/carries/v1.24/carry-commits-v1.24.log --overrides=/home/akashem/go/src/github.com/tkashem/rebase/carries/v1.24/overrides.yaml
```


comment, sha, action, clean, summary, sig, commit link, pr link

fb7cfdedf8487d917daed3f688f6aec8e0d90986


### update dependency:
- pin openshift dependencies in go.mod file
```
	github.com/openshift/api => github.com/tkashem/api bump-1.24
	github.com/openshift/client-go => github.com/tkashem/openshift-client-go bump-1.24
	github.com/openshift/apiserver-library-go => github.com/tkashem/apiserver-library-go bump-1.24
    github.com/openshift/library-go => github.com/tkashem/library-go  bump-1.24
```

- run `go mod tidy`
- run `hack/update-vendor.sh`

add a new commit with summary
```
UPSTREAM: drop: 
```

after the k8s bump merges, you'll need to wait for ART to provide the base image with new kubelet
that's where you might (but don't have to, a lot will depend on the tests itself) see failures in the k8s bump in origin
https://docs.google.com/document/d/1KM6sBcTL7xlL3TyXs5-9jX8_dAkRh2KhxVn6l6rr02g/edit#heading=h.crw5x56ksect


apiserver: add system_client=kube-{apiserver,cm,s} to apiserver_request_total



    UPSTREAM: <drop>: pin ginkgo to openshift fork at origin-4.7 branch
    
    add openshift/ginkgo in replace of go.mod
    github.com/onsi/ginkgo => github.com/openshift/ginkgo origin-4.7
    
    for the following files:
        go.mod
        staging/src/k8s.io/code-generator/go.mod
        staging/src/k8s.io/code-generator/examples/go.mod
    
    run the following comand:
    - go mod tidy
    - hack/update-vendor.sh





UPSTREAM: <drop>: update openshift dependencies

we encountered a vendor inconsistency error from staging/src/k8s.io/code-generator.
in order to fix it:
- change direcroty to 'staging/src/k8s.io/code-generator/examples'
- open go.mod and add openshift ginkgo in a replace section
  replace (
     github.com/onsi/ginkgo => github.com/openshift/ginkgo origin-4.7
  )
- go mod tidy

- change directory to 'staging/src/k8s.io/code-generator'
- open go.mod and add openshift ginkgo in a replace section
  replace (
    github.com/onsi/ginkgo => github.com/openshift/ginkgo origin-4.7
  )
- go mod tidy
- go mod vendor

- change directory to root of the repo
- bump the following openshift repos in the go.mod file
     - openshift/api
     - openshift/client-go
     - openshift/library-go
     - openshift/apiserver-library-go
     - github.com/openshift/ginkgo origin-4.7

- go mod tidy
- hack/update-vendor.sh



```
	github.com/openshift/api => github.com/tkashem/api bump-1.24
	github.com/openshift/client-go => github.com/tkashem/openshift-client-go bump-1.24
	github.com/openshift/apiserver-library-go => github.com/tkashem/apiserver-library-go bump-1.24
    github.com/openshift/library-go => github.com/tkashem/library-go  bump-1.24
    
    github.com/onsi/ginkgo => github.com/openshift/ginkgo origin-4.7
```

- go mod tidy
- make/update-vendor.sh


github.com/openshift/api 641a165d1cca377be0d7f9ba7f6f60bdbab558e8
github.com/openshift/client-go dddeb4eb20b7732dacadea7622da9b1ba13a286f
github.com/openshift/library-go 607a089b3f0ba6339bcfda71b1a0cf195f9ec6d0
github.com/openshift/apiserver-library-go 5950ef701ec5e3052c742b0f145e701930823c24




go: inconsistent vendoring in /go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/vendor/k8s.io/code-generator:
github.com/onsi/ginkgo@v4.7.0-origin.0+incompatible: is explicitly required in go.mod, but not marked as explicit in vendor/modules.txt
github.com/onsi/gomega@v1.10.1: is explicitly required in go.mod, but not marked as explicit in vendor/modules.txt
k8s.io/code-generator: is marked as replaced in vendor/modules.txt, but not replaced in go.mod

	To ignore the vendor directory, use -mod=readonly or -mod=mod.
	To sync the vendor directory, run:
		go mod vendor







go: inconsistent vendoring in /go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/vendor/k8s.io/code-generator:
k8s.io/code-generator: is marked as replaced in vendor/modules.txt, but not replaced in go.mod

	To ignore the vendor directory, use -mod=readonly or -mod=mod.
	To sync the vendor directory, run:
		go mod vendor
Running update-codegen FAILED


k8s.io/code-generator => ../code-generator




https://github.com/kubernetes/kubernetes/blob/v1.23.7-rc.0/staging/src/k8s.io/apiserver/pkg/server/options/deprecated_insecure_serving.go



UPSTREAM: <drop>: run make update

make update fails with the following error due to some go mod issue
in code-generator:
```
inconsistent vendoring in /go/{redacted}/k8s.io/code-generator
```

to work around, we do the following:

- change directory to 'staging/src/k8s.io/code-generator'
- remove the following from the 'replace' section
     k8s.io/code-generator => ../code-generator
- run 'go mod tidy'
- run 'go mod vendor'

- change directory to root
- run make update, using the following command 

docker run -it --rm -v $( pwd ):/go/src/k8s.io/kubernetes:Z \
  --workdir=/go/src/k8s.io/kubernetes \
  registry.ci.openshift.org/openshift/release:rhel-8-release-golang-1.18-openshift-4.11 \
  make update OS_RUN_WITHOUT_DOCKER=yes


UPSTREAM: <drop>: run ./hack/update-netparse-cve.sh

due to the following inconsistent vendoring error:
  go: inconsistent vendoring in /go/src/k8s.io/kubernetes/hack/tools:

we need to do the following steps:
- change directory to 'hack/tools'
- run 'go mod tidy'
- run 'go mod vendor'
- run 'update-netparse-cve.sh'

if you are inside the container, then:
OS_RUN_WITHOUT_DOCKER=yes hack/update-netparse-cve.sh




UPSTREAM: <drop>: run ./hack/update-mocks.sh

due to the following inconsistent vendoring error:
go: inconsistent vendoring in /go/src/k8s.io/kubernetes/hack/tools:

we need to do the following steps:
- change directory to 'hack/tools'
- run 'go mod tidy'
- run 'go mod vendor'
- run 'update-netparse-cve.sh'

if you are inside the container, then:
OS_RUN_WITHOUT_DOCKER=yes hack/update-netparse-cve.sh


UPSTREAM: <drop>: run ./hack/update-vendor.sh



bootstrap/containers/bootstrap-control-plane/kube-apiserver.log
start: I0427 00:00:34.932547
end:   E0427 00:30:20.176823


kube-apiserver-bea03f146248b568076c996e5ea911553643a10875531d7556cbd24d6d02911e.log
I0427 00:21:53.817432
E0427 00:30:20.176823


kube-apiserver-de1b86ce20e86157bc63fecae1db9204a023b8e5b27b8a9de770b02debc6d73a.log
start: I0427 00:00:34.932547 
end:   I0427 00:21:49.467558


```
Apr 27 00:02:17 ip-10-0-132-214 hyperkube[1443]: I0427 00:02:17.956928    1443 csi_plugin.go:1063] Failed to contact API server when waiting for CSINode publishing: csinodes.storage.k8s.io "ip-10-0-132-214.us-east-2.compute.internal" is forbidden: User "system:anonymous" cannot get resource "csinodes" in API group "storage.k8s.io" at the cluster scope
```

```
Apr 27 00:02:42 ip-10-0-132-214 hyperkube[1443]: E0427 00:02:42.278790    1443 kubelet.go:2367] "Container runtime network not ready" networkReady="NetworkReady=false reason:NetworkPluginNotReady message:Network
 plugin returns error: No CNI configuration file in /etc/kubernetes/cni/net.d/. Has your network provider started?"
Apr 27 00:02:46 ip-10-0-132-214 hyperkube[1443]: I0427 00:02:46.543132    1443 csr.go:262] certificate signing request csr-x6hrr is approved, waiting to be issued
Apr 27 00:02:46 ip-10-0-132-214 hyperkube[1443]: I0427 00:02:46.559277    1443 csr.go:258] certificate signing request csr-x6hrr is issued

```


UPSTREAM: <drop>: disable ProxyTerminatingEndpoints e2e tests




ci/prow/artifacts — Job succeeded
ci/prow/e2e-aws-csi — Job succeeded
ci/prow/e2e-aws-upgrade — Job succeeded
ci/prow/e2e-gcp-upgrade — Job succeeded
ci/prow/images — Job succeeded
ci/prow/integration — Job succeeded
ci/prow/k8s-e2e-conformance-aws — Job succeeded
ci/prow/k8s-e2e-gcp-serial — Job succeeded
ci/prow/unit — Job succeeded
ci/prow/verify — Job succeeded
