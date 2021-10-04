# Production Deployment Requirements

To deploy in production, we recommend setting up the following pieces of infrastructure: A load balancer, a kubernetes ingress controller, and a DNS pointing to the load balancer. In addition we recommend using a managed database instance (such as RDS for AWS). 


Once you have your networking infrastructure set up, apply a helm values file such as the one specified in the example file below to wire up routing through an Ingress, and set up TLS.

!!! Note
     This example uses Traefik as an Ingress controller described in detail here. To configure other ingress controllers, apply their annotations in `.Values.console.annotations`.

=== "values.yaml with activation of an enterprise license and authentication"

	```yaml
    ingress:
        enabled: true
        host: <DNS-ENTRY>
        tls:
            enabled: true
            secretName: "pach-tls"
    pachd:
        tls:
            enabled: true
            secretName: "pach-tls"
        externalService:
            enabled: true
            loadBalancerIP: <LOAD-BALANCER-IP>
    console:
        enabled: true
        annotations:
            ## annotations specific to integrate with your ingress-controller
            traefik.ingress.kubernetes.io/router.tls: "true"
            kubernetes.io/ingress.class: "traefik"
	```




