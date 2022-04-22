Pachyderm has partnered with Red Hat and put an [offering in the Red Hat
Marketplace](https://marketplace.redhat.com/en-us/products/pachyderm). That
offering is similar to Pachyderm's usual release images (including Console's
release image) but has a few differences, required by Red Hat for our inclusion
in their marketplace:

- The image is based on Red Hat's Universal Base Image, rather than scratch or
  Node or any standard, non-Red Hat images
- Extra labels have been added to the image indicating that they're maintained
  by Pachyderm
- A licenses/ directory is built into the image, containing the licenses of any
  open-source dependencies
- For the Red Hat-specific Console image, extra security checks are run

This project was done in partnership with Red Hat, who has done most of the
work of integrating Pachyderm into their data science offerings. If this image
causes the Console build/release process to fail, contact the Integrations team
(who built Pachyderm's part of this) or disable/remove the release workflow for
this image in CircleCI (this image is released on a Best Effort basis)
