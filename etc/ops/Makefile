.PHONE: \
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
	@ ansible-playbook --diff -s ansible/playbooks/dev.yml
