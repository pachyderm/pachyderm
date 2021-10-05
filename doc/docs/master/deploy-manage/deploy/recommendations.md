# Production Deployment Recommended Pre-Requisites

To deploy in production, we recommend setting up the following pieces of networking infrastructure: A load balancer, a kubernetes ingress controller, and a DNS pointing to the load balancer. In addition we recommend using a managed database instance (such as RDS for AWS). 

Once you have your networking infrastructure set up, apply a helm values file such as the one specified in the example file below to wire up routing through an Ingress, and set up TLS. We recommend using a certificate manager such as [cert-manager](https://cert-manager.io/docs/) to refresh certificates and inject them as kubernetes secrets into your cluster for the ingress and load balancer to use.

!!! Note
     This example uses [Traefik](../ingress/pach-ui-ingress/) as an Ingress controller. To configure other ingress controllers, apply their annotations in `.Values.console.annotations`.

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
            ## the example below is a provided configuration specific to traefik as an ingress-controller
            traefik.ingress.kubernetes.io/router.tls: "true"
            kubernetes.io/ingress.class: "traefik"
	```




