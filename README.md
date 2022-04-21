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

7. create a new commit with the following:
- A: hack/pin-dependency.sh openshift/api {your branch from 1}
     hack/update-vendor.sh
- B: hack/pin-dependency.sh openshift/client-go {your branch from 2}
     hack/update-vendor.sh 
- C: hack/pin-dependency.sh openshift/library-go {your branch from 3}
     hack/update-vendor.sh
- B: hack/pin-dependency.sh openshift/apiserver-library-go {your branch from 4}
     hack/update-vendor.sh

8. make update?

9. make build?


10. make test

```
$ GITHUB_AUTH_TOKEN=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 /home/akashem/go/src/github.com/tkashem/rebase/output/rebase apply --target=v1.24 --carry-commit-file=/home/akashem/go/src/github.com/tkashem/rebase/carries/v1.24/carry-commits-v1.24.log --overrides=/home/akashem/go/src/github.com/tkashem/rebase/carries/v1.24/overrides.yaml
```