# kubectl Cheat Sheet

## Cluster Context

```bash
# List all contexts
kubectl config get-contexts

# Switch context
kubectl config use-context <context>

# Set default namespace for context
kubectl config set-context --current --namespace=<namespace>

# View current context
kubectl config current-context

# View cluster info
kubectl cluster-info
```

## Resource Management

### Get Resources

```bash
# List resources (most common)
kubectl get pods                                    # all pods in current ns
kubectl get pods -n <namespace>                     # specific namespace
kubectl get pods -A                                 # all namespaces
kubectl get pods -o wide                            # extra columns (node, IP)
kubectl get pods -o yaml                            # full YAML output
kubectl get pods -o json                            # full JSON output
kubectl get pods --show-labels                      # show labels
kubectl get pods -l app=myapp                       # filter by label
kubectl get pods --field-selector status.phase=Running  # filter by field
kubectl get pods --sort-by='.status.startTime'      # sort by field

# Common resource types
kubectl get deploy,svc,pods                         # multiple types
kubectl get all                                     # common resources
kubectl get events --sort-by='.lastTimestamp'        # recent events
kubectl api-resources                               # all resource types
```

### Describe & Inspect

```bash
kubectl describe pod <pod>                          # detailed info + events
kubectl describe node <node>                        # node capacity, conditions
kubectl describe svc <service>                      # endpoints, ports
kubectl get pod <pod> -o jsonpath='{.spec.containers[*].image}'  # extract field
kubectl get pods -o custom-columns='NAME:.metadata.name,STATUS:.status.phase'
```

### Create & Apply

```bash
kubectl apply -f manifest.yaml                      # create or update
kubectl apply -f ./manifests/                       # apply all in directory
kubectl apply -k ./kustomize/                       # apply kustomization
kubectl create -f manifest.yaml                     # create only (error if exists)
kubectl create deployment nginx --image=nginx       # imperative create

# Dry run (preview without applying)
kubectl apply -f manifest.yaml --dry-run=client     # client-side validation
kubectl apply -f manifest.yaml --dry-run=server     # server-side validation
kubectl diff -f manifest.yaml                       # show what would change
```

### Edit & Patch

```bash
kubectl edit deployment <name>                      # open in $EDITOR
kubectl patch deployment <name> -p '{"spec":{"replicas":3}}'  # JSON patch
kubectl set image deployment/<name> app=image:v2    # update image
kubectl scale deployment <name> --replicas=5        # scale up/down
kubectl label pod <pod> env=prod                    # add label
kubectl annotate pod <pod> description="my pod"     # add annotation
```

### Delete

```bash
kubectl delete -f manifest.yaml                     # delete from file
kubectl delete pod <pod>                            # delete specific
kubectl delete pod <pod> --grace-period=0 --force   # force delete (stuck pods)
kubectl delete pods -l app=myapp                    # delete by label
kubectl delete namespace <namespace>                # delete ns + everything in it
```

## Debugging

### Logs

```bash
kubectl logs <pod>                                  # current logs
kubectl logs <pod> --previous                       # previous crash logs
kubectl logs <pod> -c <container>                   # specific container
kubectl logs <pod> -f                               # follow/stream
kubectl logs <pod> --tail=100                       # last 100 lines
kubectl logs <pod> --since=1h                       # last hour
kubectl logs -l app=myapp --all-containers          # all pods with label
```

### Exec & Debug

```bash
kubectl exec -it <pod> -- /bin/sh                   # shell into pod
kubectl exec <pod> -- cat /etc/config/app.conf      # run command
kubectl cp <pod>:/path/file ./local-file            # copy from pod
kubectl cp ./local-file <pod>:/path/file            # copy to pod
kubectl debug <pod> -it --image=busybox             # ephemeral debug container
kubectl debug node/<node> -it --image=busybox       # debug node
kubectl port-forward <pod> 8080:3000                # forward pod port
kubectl port-forward svc/<svc> 8080:80              # forward service port
```

### Resource Usage

```bash
kubectl top pods                                    # pod CPU/memory
kubectl top pods --sort-by=memory                   # sort by memory
kubectl top nodes                                   # node CPU/memory
kubectl top pods -A --sort-by=cpu                   # all namespaces, by CPU
```

## Deployments & Rollouts

```bash
kubectl rollout status deployment/<name>            # watch rollout
kubectl rollout history deployment/<name>           # revision history
kubectl rollout undo deployment/<name>              # rollback to previous
kubectl rollout undo deployment/<name> --to-revision=3  # rollback to specific
kubectl rollout restart deployment/<name>           # restart all pods
kubectl rollout pause deployment/<name>             # pause rollout
kubectl rollout resume deployment/<name>            # resume rollout
```

## Networking

```bash
# Services
kubectl get svc                                     # list services
kubectl get endpoints <svc>                         # check service endpoints
kubectl describe svc <svc>                          # full service info

# Ingress
kubectl get ingress                                 # list ingresses
kubectl describe ingress <name>                     # rules, backend

# DNS test
kubectl run dns-test --rm -it --image=busybox -- nslookup <svc>.<ns>.svc.cluster.local

# Network debug
kubectl run netshoot --rm -it --image=nicolaka/netshoot -- bash
# Inside: curl, dig, traceroute, tcpdump, iperf, etc.

# Network policies
kubectl get networkpolicy -A
```

## Secrets & ConfigMaps

```bash
# Create secret
kubectl create secret generic my-secret \
  --from-literal=password=s3cret \
  --from-file=ssh-key=~/.ssh/id_rsa

# Create configmap
kubectl create configmap my-config \
  --from-literal=key=value \
  --from-file=config.yaml

# View secret (base64 decoded)
kubectl get secret <name> -o jsonpath='{.data.password}' | base64 -d

# View configmap
kubectl get configmap <name> -o yaml
```

## Namespace Management

```bash
kubectl get namespaces
kubectl create namespace <name>
kubectl delete namespace <name>                     # deletes EVERYTHING in it
kubectl get all -n <namespace>
```

## RBAC

```bash
# Check permissions
kubectl auth can-i create pods                      # current user
kubectl auth can-i create pods --as=system:serviceaccount:ns:sa  # as SA
kubectl auth can-i --list                           # all permissions

# View roles
kubectl get roles,rolebindings -n <namespace>
kubectl get clusterroles,clusterrolebindings
kubectl describe role <role> -n <namespace>
```

## Useful One-Liners

```bash
# Delete all evicted pods
kubectl get pods -A --field-selector status.phase=Failed \
  -o jsonpath='{range .items[*]}{.metadata.namespace}{" "}{.metadata.name}{"\n"}{end}' | \
  xargs -L1 sh -c 'kubectl delete pod $1 -n $0'

# Get all images running in cluster
kubectl get pods -A -o jsonpath='{range .items[*]}{.spec.containers[*].image}{"\n"}{end}' | sort -u

# Find pods not running/completed
kubectl get pods -A --field-selector 'status.phase!=Running,status.phase!=Succeeded'

# Get pod resource requests/limits
kubectl get pods -o custom-columns=\
'NAME:.metadata.name,CPU_REQ:.spec.containers[0].resources.requests.cpu,MEM_REQ:.spec.containers[0].resources.requests.memory,CPU_LIM:.spec.containers[0].resources.limits.cpu,MEM_LIM:.spec.containers[0].resources.limits.memory'

# Watch pods (auto-refresh)
kubectl get pods -w

# Get events for specific pod
kubectl get events --field-selector involvedObject.name=<pod> --sort-by='.lastTimestamp'

# Force replace a resource
kubectl replace --force -f manifest.yaml

# Get all resources in a namespace
kubectl api-resources --verbs=list --namespaced -o name | \
  xargs -I {} kubectl get {} -n <namespace> --show-kind --ignore-not-found
```

## Output Formatting

```bash
# JSONPath examples
kubectl get pods -o jsonpath='{.items[*].metadata.name}'
kubectl get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\n"}{end}'

# Custom columns
kubectl get pods -o custom-columns='NAME:.metadata.name,NODE:.spec.nodeName,STATUS:.status.phase'

# Go template
kubectl get pods -o go-template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'

# Sort
kubectl get pods --sort-by='.metadata.creationTimestamp'
kubectl get pods --sort-by='.status.containerStatuses[0].restartCount'
```
