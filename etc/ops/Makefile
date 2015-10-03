.PHONE: \
	ansible-ci \
	ansible-dev \
	ansible-install

ANSIBLE_INVENTORY = $(shell pwd)/ansible/inventory/gce.py
ANSIBLE_ROLES_PATH = \
	$(shell pwd)/ansible/roles/internal:$(shell pwd)/ansible/roles/vendor

export ANSIBLE_INVENTORY ANSIBLE_ROLES_PATH

ansible-install:
	@ ansible-galaxy install \
		--ignore-errors \
		-r ansible/roles/requirements.txt \
		-p ansible/roles/vendor

ansible-dev:
	@ ansible-playbook --diff -s -u dev \
		--private-key=~/.ssh/dev_rsa \
		ansible/playbooks/dev.yml

ansible-ci:
	@ ansible-playbook --diff -s ansible/playbooks/ci.yml \
			-e "buildkite_agent_token=${BUILDKITE_AGENT_TOKEN}"

terraform-dev:
	@ terraform apply \
		-state terraform/terraform.tfstate \
		-var-file terraform/dev/terraform.tfvars \
		terraform/dev
