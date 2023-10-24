import json
import sys
import time
from typing import List

import requests


def pwChange(user: str, password: str, entrypoint: str) -> None:
    print(f"Trying to change {user}'s password", flush=True)
    while True:
        try:
            resp = requests.post(
                url=f"http://{entrypoint}/api/v1/auth/login",
                json={"username": user, "password": ""},
                verify=False,
            )
            data = json.loads(resp.text)
            requests.post(
                url=f"http://{entrypoint}/api/v1/users/{data['user']['id']}/password",
                json=password,
                headers={"Authorization": f"Bearer {data['token']}"},
                verify=False,
            )
            print(f"Successfully changed {user}'s password", flush=True)
            return
        except Exception as e:
            print(f"Encountered exception: {e}", flush=True)
            return


def checkPortAlive(entrypoint: str) -> None:
    while True:
        try:
            requests.get(
                url=f"http://{entrypoint}/api/v1/master",
                verify=False,
            )
            return
        except Exception as e:
            print(f"Encountered exception: {e}")
            continue


def getMasterAddress(
    namespace: str,
    service_name: str,
    master_port: str,
    node_port: str,
    service_host: str,
    service_port: str,
    token: str,
) -> str:

    target_service = f"determined-master-service-{service_name}"

    if node_port != "true":
        for i in range(300): # 5 minutes
            services = requests.get(
                url=f"https://{service_host}:{service_port}/api/v1/namespaces/{namespace}/services",
                headers={"Authorization": f"Bearer {token}"},
                verify=False,
            ).json()

            for svc in services["items"]:
                if target_service in svc["metadata"]["name"]:
                    status = svc["status"]["loadBalancer"]
                    if "ingress" not in svc["status"]["loadBalancer"] or status["ingress"] is None:
                        time.sleep(1) # 1 second and loop
                        break
                    if status["ingress"][0].get("hostname"):
                        # use hostname over ip address, if available
                        return f"{status['ingress'][0]['hostname']}:{master_port}"
                    else:
                        return f"{status['ingress'][0]['ip']}:{master_port}"
    else:
        services = requests.get(
            url=f"https://{service_host}:{service_port}/api/v1/namespaces/{namespace}/services",
            headers={"Authorization": f"Bearer {token}"},
            verify=False,
        ).json()
        
        for svc in services["items"]:
            if target_service in svc["metadata"]["name"]:
                #entrypoint = f"{svc['spec']['cluster_ip']}:{master_port}"
                entrypoint = f"{svc['spec']['clusterIP']}:{master_port}"
                checkPortAlive(entrypoint)

                return entrypoint
    return ""


def main(argv: List[str]) -> None:
    if len(argv) < 7:
        raise Exception("not enough args")
    if argv[7] == "":
        raise Exception("no password supplied")

    for i in range(len(argv)):
        argv[i] = argv[i].strip()
    addr = getMasterAddress(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6])
    pwChange("determined", argv[7], addr)
    pwChange("admin", argv[7], addr)


if __name__ == "__main__":
    main(sys.argv[1:])
