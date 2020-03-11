import logging as log

from sys import argv
from ruamel.yaml import YAML
from collections import Mapping
from os import path, walk, getcwd

exclude_dirs = [
    ".github",
    "docs",
    "hack",
    "kfdef",
    "plugins",
    "tests",
]

accepted_kinds = [
    "statefulset", "deployment", "daemonSet", "replicaset", "job", "cronjob"
]


def image_from_string(img_str):
    """Parse a string into an image map"""

    tag = ""
    digest = ""

    if ":" not in img_str:
        img_str = img_str + ":latest"

    if "@" in img_str:
        name, digest = img_str.rsplit("@", 1)
    else:
        name, tag = img_str.rsplit(":", 1)

    img = {"name": name, "tag": tag, "digest": digest}
    return img


def append_or_update(img_list, new_img):
    for img in img_list:
        if img["name"] == new_img["name"]:
            img["newName"] = img.get("newName", new_img["name"])
            if img.get("newTag", new_img["tag"]):
                img["newTag"] = img.get("newTag", new_img["tag"])
            if img.get("digest", new_img["digest"]):
                img["digest"] = img.get("digest", new_img["digest"])
            return
    # Image doesn't exist in transformation list
    kust_img = {"name": new_img["name"], "newName": new_img["name"]}
    if new_img["tag"]:
        kust_img["newTag"] = new_img["tag"]
    if new_img["digest"]:
        kust_img["digest"] = new_img["digest"]
    img_list.append(kust_img)


def scan_kustomization_for_images(kust_dir):
    """Scan kustomization folder and produce a list of images

    Args:
      kust_dir (str): Path where the kustomize application resides.
    """

    yaml = YAML()
    yaml.block_seq_indent = 0

    # Load kustomization
    with open(path.join(kust_dir, "kustomization.yaml")) as f:
        try:
            kustomization = yaml.load(f)
        except Exception as e:
            log.error("Error loading kustomization in %s: %s", kust_dir, e)
            raise (e)

    # Get current image list from kustomization
    img_list = kustomization.get("images", [])

    # Get local resource files
    (_, _, filenames) = next(walk(kust_dir))
    filenames = [
        filename for filename in filenames
        if filename != "kustomization.yaml" and filename != "params.yaml" and
        filename.endswith(".yaml")
    ]

    for filename in filenames:
        with open(path.join(kust_dir, filename)) as f:
            resources = list(yaml.load_all(f))
        for r in resources:
            if not isinstance(r, Mapping):
                continue
            if r.get("kind", "").lower() in accepted_kinds:
                try:
                    containers = r["spec"]["template"]["spec"]["containers"]
                except KeyError:
                    continue
                for c in containers:
                    try:
                        img_str = c["image"]
                    except KeyError:
                        continue
                    new_img = image_from_string(img_str)
                    append_or_update(img_list, new_img)

    if img_list:
        kustomization["images"] = img_list
        with open(path.join(kust_dir, "kustomization.yaml"), "w") as f:
            yaml.dump(kustomization, f)


def check_kustomize_dir(d):
    (curr, folders, _) = next(walk(d))
    if len(folders) == 1 and "base" in folders:
        return True, [path.join(curr, "base")]
    if len(folders) == 2 and "base" in folders and "overlays" in folders:
        (_, folders, _) = next(walk(path.join(curr, "overlays")))
        return True, [path.join(curr, "base")] + [path.join(curr, "overlays", f) for f in folders]
    return False, []


def get_kustomization_dirs(root):

    def helper(root, dirs):
        (curr, folders, _) = next(walk(root))
        (is_kustomize_dir, kust_dirs) = check_kustomize_dir(curr)
        if is_kustomize_dir:
            dirs.extend(kust_dirs)
            return
        for f in folders:
            if f in exclude_dirs:
                continue
            helper(path.join(curr, f), dirs)

    res = []
    helper(root, res)
    return res


if __name__ == "__main__":
    log.basicConfig(
        level=log.INFO,
        format=('%(levelname)s|%(asctime)s'
                '|%(pathname)s|%(lineno)d| %(message)s'),
        datefmt='%Y-%m-%dT%H:%M:%S',
    )
    log.getLogger().setLevel(log.INFO)

    root_dir = getcwd() if len(argv) < 2 else argv[1]
    for d in get_kustomization_dirs(root_dir):
        scan_kustomization_for_images(d)