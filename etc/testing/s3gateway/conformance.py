#!/usr/bin/env python3

import os
import re
import sys
import glob
import time
import datetime
import threading
import subprocess
import http.client
import collections

TEST_ROOT = os.path.join("etc", "testing", "s3gateway")
RUNS_ROOT = os.path.join(TEST_ROOT, "runs")
RAN_PATTERN = re.compile(r"Ran (\d+) tests in [\d\.]+s")
FAILED_PATTERN = re.compile(r"FAILED \(SKIP=(\d+), errors=(\d+), failures=(\d+)\)")

ERROR_PATTERN = re.compile(r"ERROR: (s3tests\..+)")

TRACEBACK_PREFIXES = [
    "Traceback (most recent call last):",
    "----------------------------------------------------------------------",
    "  ",
]

BLACKLISTED_TESTS = [
    "s3tests.functional.test_s3.test_bucket_list_return_data_versioning",
    "s3tests.functional.test_s3.test_bucket_list_marker_versioning",
    "s3tests.functional.test_s3.test_bucket_list_objects_anonymous",
    "s3tests.functional.test_s3.test_post_object_anonymous_request",
    "s3tests.functional.test_s3.test_post_object_authenticated_request_bad_access_key",
    "s3tests.functional.test_s3.test_post_object_set_success_code",
    "s3tests.functional.test_s3.test_post_object_set_invalid_success_code",
    "s3tests.functional.test_s3.test_post_object_success_redirect_action",
    "s3tests.functional.test_s3.test_object_raw_put_write_access",
    "s3tests.functional.test_s3.test_bucket_acl_default",
    "s3tests.functional.test_s3.test_bucket_acl_canned_during_create",
    "s3tests.functional.test_s3.test_bucket_acl_canned",
    "s3tests.functional.test_s3.test_bucket_acl_canned_publicreadwrite",
    "s3tests.functional.test_s3.test_bucket_acl_canned_authenticatedread",
    "s3tests.functional.test_s3.test_object_acl_canned_bucketownerread",
    "s3tests.functional.test_s3.test_object_acl_canned_bucketownerfullcontrol",
    "s3tests.functional.test_s3.test_object_acl_full_control_verify_owner",
    "s3tests.functional.test_s3.test_object_acl_full_control_verify_attributes",
    "s3tests.functional.test_s3.test_bucket_acl_canned_private_to_private",
    "s3tests.functional.test_s3.test_bucket_acl_xml_fullcontrol",
    "s3tests.functional.test_s3.test_bucket_acl_xml_write",
    "s3tests.functional.test_s3.test_bucket_acl_xml_writeacp",
    "s3tests.functional.test_s3.test_bucket_acl_xml_read",
    "s3tests.functional.test_s3.test_bucket_acl_xml_readacp",
    "s3tests.functional.test_s3.test_bucket_acl_grant_userid_fullcontrol",
    "s3tests.functional.test_s3.test_bucket_acl_grant_userid_read",
    "s3tests.functional.test_s3.test_bucket_acl_grant_userid_readacp",
    "s3tests.functional.test_s3.test_bucket_acl_grant_userid_write",
    "s3tests.functional.test_s3.test_bucket_acl_grant_userid_writeacp",
    "s3tests.functional.test_s3.test_bucket_acl_grant_nonexist_user",
    "s3tests.functional.test_s3.test_bucket_header_acl_grants",
    "s3tests.functional.test_s3.test_bucket_acl_grant_email",
    "s3tests.functional.test_s3.test_bucket_acl_grant_email_notexist",
    "s3tests.functional.test_s3.test_bucket_acl_revoke_all",
    "s3tests.functional.test_s3.test_logging_toggle",
    "s3tests.functional.test_s3.test_access_bucket_private_object_private",
    "s3tests.functional.test_s3.test_access_bucket_private_object_publicread",
    "s3tests.functional.test_s3.test_access_bucket_private_object_publicreadwrite",
    "s3tests.functional.test_s3.test_access_bucket_publicread_object_private",
    "s3tests.functional.test_s3.test_access_bucket_publicread_object_publicread",
    "s3tests.functional.test_s3.test_access_bucket_publicread_object_publicreadwrite",
    "s3tests.functional.test_s3.test_access_bucket_publicreadwrite_object_private",
    "s3tests.functional.test_s3.test_access_bucket_publicreadwrite_object_publicread",
    "s3tests.functional.test_s3.test_access_bucket_publicreadwrite_object_publicreadwrite",
    "s3tests.functional.test_s3.test_object_copy_versioned_bucket",
    "s3tests.functional.test_s3.test_object_copy_versioning_multipart_upload",
    "s3tests.functional.test_s3.test_multipart_upload_empty",
    "s3tests.functional.test_s3.test_multipart_upload_small",
    "s3tests.functional.test_s3.test_multipart_upload",
    "s3tests.functional.test_s3.test_multipart_copy_versioned",
    "s3tests.functional.test_s3.test_multipart_upload_resend_part",
    "s3tests.functional.test_s3.test_multipart_upload_multiple_sizes",
    "s3tests.functional.test_s3.test_multipart_upload_size_too_small",
    "s3tests.functional.test_s3.test_multipart_upload_contents",
    "s3tests.functional.test_s3.test_abort_multipart_upload",
    "s3tests.functional.test_s3.test_list_multipart_upload",
    "s3tests.functional.test_s3.test_multipart_upload_missing_part",
    "s3tests.functional.test_s3.test_multipart_upload_incorrect_etag",
    "s3tests.functional.test_s3.test_bucket_acls_changes_persistent",
    "s3tests.functional.test_s3.test_stress_bucket_acls_changes",
    "s3tests.functional.test_s3.test_cors_origin_response",
    "s3tests.functional.test_s3.test_cors_origin_wildcard",
    "s3tests.functional.test_s3.test_cors_header_option",
    "s3tests.functional.test_s3.test_multipart_resend_first_finishes_last",
    "s3tests.functional.test_s3.test_versioning_bucket_create_suspend",
    "s3tests.functional.test_s3.test_versioning_obj_plain_null_version_removal",
    "s3tests.functional.test_s3.test_versioning_obj_plain_null_version_overwrite",
    "s3tests.functional.test_s3.test_versioning_obj_plain_null_version_overwrite_suspended",
    "s3tests.functional.test_s3.test_versioning_obj_suspend_versions",
    "s3tests.functional.test_s3.test_versioning_obj_suspend_versions_simple",
    "s3tests.functional.test_s3.test_versioning_obj_create_versions_remove_all",
    "s3tests.functional.test_s3.test_versioning_obj_create_versions_remove_special_names",
    "s3tests.functional.test_s3.test_versioning_obj_create_overwrite_multipart",
    "s3tests.functional.test_s3.test_versioning_obj_list_marker",
    "s3tests.functional.test_s3.test_versioning_copy_obj_version",
    "s3tests.functional.test_s3.test_versioning_multi_object_delete",
    "s3tests.functional.test_s3.test_versioning_multi_object_delete_with_marker",
    "s3tests.functional.test_s3.test_versioning_multi_object_delete_with_marker_create",
    "s3tests.functional.test_s3.test_versioned_object_acl",
    "s3tests.functional.test_s3.test_versioned_object_acl_no_version_specified",
    "s3tests.functional.test_s3.test_versioned_concurrent_object_create_concurrent_remove",
    "s3tests.functional.test_s3.test_versioned_concurrent_object_create_and_remove",
    "s3tests.functional.test_s3.test_lifecycle_set",
    "s3tests.functional.test_s3.test_lifecycle_get",
    "s3tests.functional.test_s3.test_lifecycle_get_no_id",
    "s3tests.functional.test_s3.test_lifecycle_expiration",
    "s3tests.functional.test_s3.test_lifecycle_noncur_expiration",
    "s3tests.functional.test_s3.test_lifecycle_deletemarker_expiration",
    "s3tests.functional.test_s3.test_lifecycle_multipart_expiration",
    "s3tests.functional.test_s3.test_encryption_sse_c_multipart_upload",
    "s3tests.functional.test_s3.test_encryption_sse_c_multipart_bad_download",
    "s3tests.functional.test_s3.test_sse_kms_multipart_upload",
    "s3tests.functional.test_s3.test_sse_kms_multipart_invalid_chunks_1",
    "s3tests.functional.test_s3.test_sse_kms_multipart_invalid_chunks_2",
    "s3tests.functional.test_s3.test_post_object_tags_anonymous_request",
    "s3tests.functional.test_s3.test_versioning_bucket_atomic_upload_return_version_id",
    "s3tests.functional.test_s3.test_versioning_bucket_multipart_upload_return_version_id",
    "s3tests.functional.test_s3.test_bucket_policy_put_obj_acl",
    "s3tests.functional.test_s3.test_bucket_policy_put_obj_grant",
    "s3tests.functional.test_s3.test_bucket_policy_put_obj_enc",
    "s3tests.functional.test_s3.test_bucket_policy_put_obj_request_obj_tag",
    "s3tests.functional.test_s3_website",
    "s3tests.functional.test_headers.test_object_acl_create_contentlength_none",
    "s3tests.functional.test_s3.test_bucket_list_return_data",
    "s3tests.functional.test_s3.test_multi_object_delete",
    "s3tests.functional.test_s3.test_object_raw_get",
    "s3tests.functional.test_s3.test_object_raw_get_bucket_gone",
    "s3tests.functional.test_s3.test_object_raw_get_object_gone",
    "s3tests.functional.test_s3.test_object_raw_get_bucket_acl",
    "s3tests.functional.test_s3.test_object_raw_get_object_acl",
    "s3tests.functional.test_s3.test_object_raw_authenticated",
    "s3tests.functional.test_s3.test_object_raw_response_headers",
    "s3tests.functional.test_s3.test_object_raw_authenticated_bucket_acl",
    "s3tests.functional.test_s3.test_object_raw_authenticated_object_acl",
    "s3tests.functional.test_s3.test_object_raw_authenticated_bucket_gone",
    "s3tests.functional.test_s3.test_object_raw_authenticated_object_gone",
    "s3tests.functional.test_s3.test_object_acl_default",
    "s3tests.functional.test_s3.test_object_acl_canned_during_create",
    "s3tests.functional.test_s3.test_object_acl_canned",
    "s3tests.functional.test_s3.test_object_acl_canned_publicreadwrite",
    "s3tests.functional.test_s3.test_object_acl_canned_authenticatedread",
    "s3tests.functional.test_s3.test_object_acl_xml",
    "s3tests.functional.test_s3.test_object_acl_xml_write",
    "s3tests.functional.test_s3.test_object_acl_xml_writeacp",
    "s3tests.functional.test_s3.test_object_acl_xml_read",
    "s3tests.functional.test_s3.test_object_acl_xml_readacp",
    "s3tests.functional.test_s3.test_bucket_acl_no_grants",
    "s3tests.functional.test_s3.test_object_header_acl_grants",
    "s3tests.functional.test_s3.test_object_set_valid_acl",
    "s3tests.functional.test_s3.test_object_giveaway",
    "s3tests.functional.test_s3.test_bucket_create_special_key_names",
    "s3tests.functional.test_s3.test_object_copy_zero_size",
    "s3tests.functional.test_s3.test_object_copy_same_bucket",
    "s3tests.functional.test_s3.test_object_copy_verify_contenttype",
    "s3tests.functional.test_s3.test_object_copy_to_itself_with_metadata",
    "s3tests.functional.test_s3.test_object_copy_diff_bucket",
    "s3tests.functional.test_s3.test_object_copy_not_owned_bucket",
    "s3tests.functional.test_s3.test_object_copy_not_owned_object_bucket",
    "s3tests.functional.test_s3.test_object_copy_canned_acl",
    "s3tests.functional.test_s3.test_object_copy_retaining_metadata",
    "s3tests.functional.test_s3.test_object_copy_replacing_metadata",
    "s3tests.functional.test_s3.test_multipart_copy_small",
    "s3tests.functional.test_s3.test_multipart_copy_invalid_range",
    "s3tests.functional.test_s3.test_multipart_copy_without_range",
    "s3tests.functional.test_s3.test_multipart_copy_special_names",
    "s3tests.functional.test_s3.test_multipart_copy_multiple_sizes",
    "s3tests.functional.test_s3.test_multipart_upload_overwrite_existing_object",
    "s3tests.functional.test_s3.test_atomic_multipart_upload_write",
    "s3tests.functional.test_s3.test_versioning_obj_create_read_remove",
    "s3tests.functional.test_s3.test_versioning_obj_create_read_remove_head",
    "s3tests.functional.test_s3.test_bucket_policy",
    "s3tests.functional.test_s3.test_bucket_policy_acl",
    "s3tests.functional.test_s3.test_bucket_policy_different_tenant",
    "s3tests.functional.test_s3.test_bucket_policy_another_bucket",
    "s3tests.functional.test_s3.test_bucket_policy_set_condition_operator_end_with_IfExists",
    "s3tests.functional.test_s3.test_bucket_policy_list_bucket_with_prefix",
    "s3tests.functional.test_s3.test_bucket_policy_list_bucket_with_maxkeys",
    "s3tests.functional.test_s3.test_bucket_policy_list_bucket_with_delimiter",
    "s3tests.functional.test_s3.test_bucket_policy_list_put_bucket_acl_canned_acl",
    "s3tests.functional.test_s3.test_bucket_policy_list_put_bucket_acl_grants",
    "s3tests.functional.test_s3.test_get_tags_acl_public",
    "s3tests.functional.test_s3.test_put_tags_acl_public",
    "s3tests.functional.test_s3.test_delete_tags_obj_public",
    "s3tests.functional.test_s3.test_bucket_policy_get_obj_existing_tag",
    "s3tests.functional.test_s3.test_bucket_policy_get_obj_tagging_existing_tag",
    "s3tests.functional.test_s3.test_bucket_policy_put_obj_tagging_existing_tag",
    "s3tests.functional.test_s3.test_bucket_policy_put_obj_copy_source",
    "s3tests.functional.test_s3.test_bucket_policy_put_obj_copy_source_meta",
    "s3tests.functional.test_s3.test_bucket_policy_get_obj_acl_existing_tag",
]

class Gateway:
    def target(self):
        self.proc = subprocess.Popen(["pachctl", "s3gateway", "-v"])

    def __enter__(self):
        t = threading.Thread(target=self.target)
        t.daemon = True
        t.start()

    def __exit__(self, type, value, traceback):
        if hasattr(self, "proc"):
            self.proc.kill()

def compute_stats(filename):
    ran = 0
    skipped = 0
    errored = 0
    failed = 0

    with open(filename, "r") as f:
        for line in f:
            match = RAN_PATTERN.match(line)
            if match:
                ran = int(match.groups()[0])
                continue

            match = FAILED_PATTERN.match(line)
            if match:
                skipped = int(match.groups()[0])
                errored = int(match.groups()[1])
                failed = int(match.groups()[2])

    if ran != 0 and skipped != 0 and errored != 0 and failed != 0:
        return (ran - skipped - errored - failed, ran - skipped)
    else:
        return (0, 0)

def run_nosetests(*args, **kwargs):
    args = [os.path.join("virtualenv", "bin", "nosetests"), *args]
    env = dict(os.environ)
    env["S3TEST_CONF"] = os.path.join("..", "s3gateway.conf")
    if "env" in kwargs:
        env.update(kwargs.pop("env"))
    cwd = os.path.join(TEST_ROOT, "s3-tests")
    proc = subprocess.run(args, env=env, cwd=cwd, **kwargs)
    print("Test run exited with {}".format(proc.returncode))

def print_failures():
    log_files = sorted(glob.glob(os.path.join(RUNS_ROOT, "*.txt")))

    if len(log_files) == 0:
        print("No log files found", file=sys.stderr)
        return 1

    old_stats = None
    if len(log_files) > 1:
        old_stats = compute_stats(log_files[-2])

    filepath = log_files[-1]
    stats = compute_stats(filepath)

    if old_stats:
        print("Overall results: {}/{} (vs last run: {}/{})".format(*stats, *old_stats))
    else:
        print("Overall results: {}/{}".format(*stats))

    failing_test = None
    causes = collections.defaultdict(lambda: [])
    with open(filepath, "r") as f:
        for line in f:
            line = line.rstrip()

            if failing_test is None:
                match = ERROR_PATTERN.match(line)
                if match is not None:
                    failing_test = match.groups()[0]
            else:
                if not any(line.startswith(p) for p in TRACEBACK_PREFIXES):
                    causes[line].append(failing_test)
                    failing_test = None

    causes = sorted(causes.items(), key=lambda i: len(i[1]), reverse=True)
    for (cause_name, failing_tests) in causes:
        print("{}:".format(cause_name))
        for failing_test in failing_tests:
            print("- {}".format(failing_test))

    return 0

def main():
    if len(sys.argv) == 2 and sys.argv[1] == "--no-run":
        sys.exit(print_failures())
    else:
        tests = sys.argv[1:]
        output = subprocess.run("ps -ef | grep pachctl | grep -v grep", shell=True, stdout=subprocess.PIPE).stdout

        if len(output.strip()) > 0:
            print("It looks like `pachctl` is already running. Please kill it before running conformance tests.", file=sys.stderr)
            sys.exit(1)

        proc = subprocess.run("yes | pachctl delete-all", shell=True)
        if proc.returncode != 0:
            raise Exception("bad exit code: {}".format(proc.returncode))

        with Gateway():
            for _ in range(10):
                conn = http.client.HTTPConnection("localhost:30600")

                try:
                    conn.request("GET", "/")
                    response = conn.getresponse()
                    if response.status == 200:
                        break
                except ConnectionRefusedError:
                    pass

                conn.close()
                print("Waiting for s3gateway...")
                time.sleep(1)
            else:
                print("s3gateway did not start", file=sys.stderr)
                sys.exit(1)

            if tests:
                print("Running tests: {}".format(", ".join(tests)))

                # In some places, nose and its plugins expect tests to
                # specified as testmodule.testname, but here, it's expected to
                # be testmodule:testname. This replaces the last . with a : so
                # that the testmodule.testname format can be used everywhere,
                # including here.
                tests = [t if ":" in t else ":".join(t.rsplit(".", 1)) for t in tests]

                run_nosetests(*tests)
            else:
                print("Running all tests")

                # This uses the `nose-exclude` plugin to exclude tests for
                # unsupported features. Note that `nosetest` does have a
                # built-in way of excluding tests, but it only seems to match
                # on top-level modules, rather than on specific tests.
                extra_env = {
                    "NOSE_EXCLUDE_TESTS": ";".join(BLACKLISTED_TESTS)
                }

                filepath = os.path.join(RUNS_ROOT, datetime.datetime.now().strftime("%Y-%m-%d-%H-%M-%S.txt"))
                with open(filepath, "w") as f:
                    run_nosetests("-a" "!fails_on_aws", env=extra_env, stderr=f)

                sys.exit(print_failures())
                
if __name__ == "__main__":
    main()
