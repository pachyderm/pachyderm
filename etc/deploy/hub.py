#!/usr/bin/env python3

import os
import re
import json
import time
import asyncio
import secrets
import argparse
import http.client

# a way to import util.py without adding it to the PYTHONPATH :o
import pathlib
import importlib.util
util_spec = importlib.util.spec_from_file_location("util", pathlib.Path(__file__).parent.parent / "util.py")
util = importlib.util.module_from_spec(util_spec)
util_spec.loader.exec_module(util)

API_KEY = os.environ["PACH_HUB_API_KEY"]
ORG_ID = os.environ["PACH_HUB_ORG_ID"]

class HubApiError(Exception):
    """Encapsulates a hub API error response"""

    def __init__(self, errors):
        def get_message(error):
            try:
                return f"{error['title']}: {error['detail']}"
            except KeyError:
                return json.dumps(error)

        if len(errors) > 1:
            message = ["multiple errors:"]
            for error in errors:
                message.append(f"- {get_message(error)}")
            message = "\n".join(message)
        else:
            message = get_message(errors[0])

        super().__init__(message)
        self.errors = errors

def request(method, endpoint, body=None):
    headers = {
        "Authorization": f"Api-Key {API_KEY}",
    }
    if body is not None:
        body = json.dumps({
            "data": {
                "attributes": body,
            }
        })

    conn = http.client.HTTPSConnection(f"hub.pachyderm.com")
    conn.request(method, f"/api/v1{endpoint}", headers=headers, body=body)
    response = conn.getresponse()
    j = json.load(response)

    if "errors" in j:
        raise HubApiError(j["errors"])
    
    return j["data"]

async def push_image(src):
    dst = src.replace(":local", ":" + (await util.get_client_version()))
    await util.run("docker", "tag", src, dst)
    await util.run("docker", "push", dst)

async def main():
    parser = argparse.ArgumentParser(description="Deploys or destroys a hub cluster.")
    parser.add_argument("action", default="deploy", help="What action to execute")
    parser.add_argument("--cluster-name", default="", help="Name of the cluster")
    parser.add_argument("--skip-build", action="store_true", help="Skip re-build of images")
    parser.add_argument("--expire-hours", type=int, default=0, help="How many hours from now to expire the cluster, defaults to no expiration")
    args = parser.parse_args()

    if args.action == "deploy":
        procs = [util.run("make", "install")]
        if not args.skip_build:
            procs.append(util.run("make", "docker-build"))
        await asyncio.gather(*procs)

        if not args.skip_build:
            await asyncio.gather(
                push_image("pachyderm/pachd:local"),
                push_image("pachyderm/worker:local"),
            )
            # wait a few seconds, as otherwise the hub API may not recognized
            # the just-pushed docker images
            await asyncio.sleep(5)

        response = request("POST", f"/organizations/{ORG_ID}/pachs", body={
            "name": args.cluster_name or f"sandbox-{secrets.token_hex(4)}",
            "pachVersion": await util.get_client_version(),
            "expireAt": time.time() + args.expire_hours * 60 * 60 if args.expire_hours > 0 else None,
        })

        cluster_name = response["attributes"]["name"]
        gke_name = response["attributes"]["gkeName"]
        pach_id = response["id"]

        await util.run("pachctl", "config", "set", "context", cluster_name, stdin=json.dumps({
            "source": 2,
            "pachd_address": f"grpcs://{gke_name}.clusters.pachyderm.io:31400",
        }))

        await util.run("pachctl", "config", "set", "active-context", cluster_name)

        # wait up to 30min for the cluster to work
        await util.retry(util.ping, attempts=3*60, sleep=10)

        async def get_otp():
            response = request("GET", f"/organizations/{ORG_ID}/pachs/{pach_id}/otps")
            return response["attributes"]["otp"]
        otp = await util.retry(get_otp, sleep=5)

        await util.run("pachctl", "auth", "login", "--one-time-password", stdin=f"{otp}\n")
    elif args.action == "destroy":
        for pach in request("GET", f"/organizations/{ORG_ID}/pachs?limit=100"):
            if pach["attributes"]["name"] == args.cluster_name:
                request("DELETE", f"/organizations/{ORG_ID}/pachs/{pach['id']}")
                await util.run("pachctl", "config", "delete", "context", pach["attributes"]["name"], raise_on_error=False)
    elif args.action == "clear":
        for pach in request("GET", f"/organizations/{ORG_ID}/pachs?limit=100"):
            print(pach["attributes"]["name"])
            request("DELETE", f"/organizations/{ORG_ID}/pachs/{pach['id']}")
            await util.run("pachctl", "config", "delete", "context", pach["attributes"]["name"], raise_on_error=False)
    else:
        raise Exception(f"unknown action: {args.action}")

if __name__ == "__main__":
    asyncio.run(main())
