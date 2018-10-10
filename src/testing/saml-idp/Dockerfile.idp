# This Dockerfile creates the image that is actually run in Kubernetes as part
# of our SAML tests. See Dockerfile.build for more info on how the static binary
# is built.
FROM scratch
# Copy statically-built IdP server into image
COPY ["./saml-idp", "./"]
EXPOSE 80
CMD [ "/saml-idp" ]
