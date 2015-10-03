.PHONE: \
	ansible-ci \
	ansible-dev \
	ansible-roles \
	setup-dev-keys \
	terraform-dev

ANSIBLE_INVENTORY = $(shell pwd)/ansible/inventory/gce.py
ANSIBLE_ROLES_PATH = \
	$(shell pwd)/ansible/roles/internal:$(shell pwd)/ansible/roles/vendor

export ANSIBLE_INVENTORY ANSIBLE_ROLES_PATH

ansible-ci:
	@ ansible-playbook --diff -s ansible/playbooks/ci.yml \
			-e "buildkite_agent_token=${BUILDKITE_AGENT_TOKEN}"

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

terraform-dev:
	@ terraform apply \
		-state terraform/terraform.tfstate \
		-var-file terraform/dev/terraform.tfvars \
		terraform/dev
