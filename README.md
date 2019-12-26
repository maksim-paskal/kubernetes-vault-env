kubectl -n vault exec -it vault-0 sh

vault auth enable kubernetes

vault write auth/kubernetes/config \
   token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
   kubernetes_host=https://${KUBERNETES_PORT_443_TCP_ADDR}:443 \
   kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

vault write auth/kubernetes/role/default \
   bound_service_account_names=* \
   bound_service_account_namespaces=demo \
   policies=default \
   ttl=1h