# P2P Validator test

* Deploy Cosmos testnet chain;
* Max server security level (check with lynis > 80);
* Collect chain metrics with prometheus exporter writen in Go lang;

Ansible playbooks were written for each of those task, here they are:
```
# Install and init cosmos chain testnet
ansible-playbook cosmos.yaml -i hosts

# Install lynis and bring the server to the required security level
ansible-playbook lynis.yaml -i hosts

# Set up metrics exporter
ansible-playbook exporter.yaml -i hosts
```
