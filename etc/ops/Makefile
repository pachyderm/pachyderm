.PHONE: \
	ansible-ci \
	ansible-dev \
	ansible-roles \
	setup-dev-keys \
	terraform-ci \
	terraform-dev

ANSIBLE_INVENTORY = $(shell pwd)/ansible/inventory/gce.py
ANSIBLE_ROLES_PATH = \
	$(shell pwd)/ansible/roles/internal:$(shell pwd)/ansible/roles/vendor
ANSIBLE_HOST_KEY_CHECKING = False

export ANSIBLE_INVENTORY \
	ANSIBLE_ROLES_PATH \
	ANSIBLE_HOST_KEY_CHECKING

ansible-ci:
	@ ansible-playbook --diff -s -u dev \
		--private-key=~/.ssh/dev_rsa \
		ansible/playbooks/ci.yml

ansible-dev:
	@ ansible-playbook --diff -s -u dev \
		--private-key=~/.ssh/dev_rsa \
		ansible/playbooks/dev.yml

ansible-roles:
	@ ansible-galaxy install \
		--ignore-errors \
		-r ansible/roles/requirements.txt \
		-p ansible/roles/vendor

setup-dev-keys:
	@ gcloud compute project-info add-metadata \
		--metadata-from-file sshKeys=terraform/dev/keys.pub

terraform-ci:
	@ terraform apply \
		-state terraform/state/ci.tfstate \
		-var-file terraform/ci/terraform.tfvars \
		terraform/ci

terraform-dev:
	@ terraform apply \
		-state terraform/state/dev.tfstate \
		-var-file terraform/dev/terraform.tfvars \
		terraform/dev
