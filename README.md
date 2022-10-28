# stepjob operator
Each automated step is assembled based on the container.It is convenient to do some automatic operation and cicd operation of middleware tools later.


# Build
* git clone https://github.com/vega-punk/stepjob-operator
* cd stepjob-operator
* kubectl apply -k config/crd/
* kubectl apply -k config/rbac/
* kubectl apply -k config/manager/
* kubectl get pod -nsystem