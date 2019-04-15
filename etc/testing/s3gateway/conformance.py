#!/usr/bin/env python3

import os
import re
import sys
import glob
import time
import argparse
import datetime
import threading
import subprocess
import http.client
import collections

TEST_ROOT = os.path.join("etc", "testing", "s3gateway")
RUNS_ROOT = os.path.join(TEST_ROOT, "runs")
RAN_PATTERN = re.compile(r"^Ran (\d+) tests in [\d\.]+s")
FAILED_PATTERN = re.compile(r"^FAILED \((SKIP=(\d+))?(, errors=(\d+))?(, failures=(\d+))?\)")

ERROR_PATTERN = re.compile(r"^(FAIL|ERROR): (.+)")

TRACEBACK_PREFIXES = [
    "Traceback (most recent call last):",
    "----------------------------------------------------------------------",
    "  ",
]

# Ignores for boto2 functional tests
BLACKLISTED_BOTO2_FUNCTIONAL_TESTS = [
    # These tests use functionality that we don't support
    "test_s3.test_bucket_list_return_data_versioning",
    "test_s3.test_bucket_list_marker_versioning",
    "test_s3.test_bucket_list_objects_anonymous",
    "test_s3.test_post_object_anonymous_request",
    "test_s3.test_post_object_authenticated_request_bad_access_key",
    "test_s3.test_post_object_set_success_code",
    "test_s3.test_post_object_set_invalid_success_code",
    "test_s3.test_post_object_success_redirect_action",
    "test_s3.test_object_raw_put_write_access",
    "test_s3.test_bucket_acl_default",
    "test_s3.test_bucket_acl_canned_during_create",
    "test_s3.test_bucket_acl_canned",
    "test_s3.test_bucket_acl_canned_publicreadwrite",
    "test_s3.test_bucket_acl_canned_authenticatedread",
    "test_s3.test_object_acl_canned_bucketownerread",
    "test_s3.test_object_acl_canned_bucketownerfullcontrol",
    "test_s3.test_object_acl_full_control_verify_owner",
    "test_s3.test_object_acl_full_control_verify_attributes",
    "test_s3.test_bucket_acl_canned_private_to_private",
    "test_s3.test_bucket_acl_xml_fullcontrol",
    "test_s3.test_bucket_acl_xml_write",
    "test_s3.test_bucket_acl_xml_writeacp",
    "test_s3.test_bucket_acl_xml_read",
    "test_s3.test_bucket_acl_xml_readacp",
    "test_s3.test_bucket_acl_grant_userid_fullcontrol",
    "test_s3.test_bucket_acl_grant_userid_read",
    "test_s3.test_bucket_acl_grant_userid_readacp",
    "test_s3.test_bucket_acl_grant_userid_write",
    "test_s3.test_bucket_acl_grant_userid_writeacp",
    "test_s3.test_bucket_acl_grant_nonexist_user",
    "test_s3.test_bucket_header_acl_grants",
    "test_s3.test_bucket_acl_grant_email",
    "test_s3.test_bucket_acl_grant_email_notexist",
    "test_s3.test_bucket_acl_revoke_all",
    "test_s3.test_logging_toggle",
    "test_s3.test_access_bucket_private_object_private",
    "test_s3.test_access_bucket_private_object_publicread",
    "test_s3.test_access_bucket_private_object_publicreadwrite",
    "test_s3.test_access_bucket_publicread_object_private",
    "test_s3.test_access_bucket_publicread_object_publicread",
    "test_s3.test_access_bucket_publicread_object_publicreadwrite",
    "test_s3.test_access_bucket_publicreadwrite_object_private",
    "test_s3.test_access_bucket_publicreadwrite_object_publicread",
    "test_s3.test_access_bucket_publicreadwrite_object_publicreadwrite",
    "test_s3.test_object_copy_versioned_bucket",
    "test_s3.test_object_copy_versioning_multipart_upload",
    "test_s3.test_multipart_upload_empty",
    "test_s3.test_multipart_upload_small",
    "test_s3.test_multipart_upload",
    "test_s3.test_multipart_copy_versioned",
    "test_s3.test_multipart_upload_resend_part",
    "test_s3.test_multipart_upload_multiple_sizes",
    "test_s3.test_multipart_upload_size_too_small",
    "test_s3.test_multipart_upload_contents",
    "test_s3.test_abort_multipart_upload",
    "test_s3.test_list_multipart_upload",
    "test_s3.test_multipart_upload_missing_part",
    "test_s3.test_multipart_upload_incorrect_etag",
    "test_s3.test_bucket_acls_changes_persistent",
    "test_s3.test_stress_bucket_acls_changes",
    "test_s3.test_cors_origin_response",
    "test_s3.test_cors_origin_wildcard",
    "test_s3.test_cors_header_option",
    "test_s3.test_multipart_resend_first_finishes_last",
    "test_s3.test_versioning_bucket_create_suspend",
    "test_s3.test_versioning_obj_plain_null_version_removal",
    "test_s3.test_versioning_obj_plain_null_version_overwrite",
    "test_s3.test_versioning_obj_plain_null_version_overwrite_suspended",
    "test_s3.test_versioning_obj_suspend_versions",
    "test_s3.test_versioning_obj_suspend_versions_simple",
    "test_s3.test_versioning_obj_create_versions_remove_all",
    "test_s3.test_versioning_obj_create_versions_remove_special_names",
    "test_s3.test_versioning_obj_create_overwrite_multipart",
    "test_s3.test_versioning_obj_list_marker",
    "test_s3.test_versioning_copy_obj_version",
    "test_s3.test_versioning_multi_object_delete",
    "test_s3.test_versioning_multi_object_delete_with_marker",
    "test_s3.test_versioning_multi_object_delete_with_marker_create",
    "test_s3.test_versioned_object_acl",
    "test_s3.test_versioned_object_acl_no_version_specified",
    "test_s3.test_versioned_concurrent_object_create_concurrent_remove",
    "test_s3.test_versioned_concurrent_object_create_and_remove",
    "test_s3.test_lifecycle_set",
    "test_s3.test_lifecycle_get",
    "test_s3.test_lifecycle_get_no_id",
    "test_s3.test_lifecycle_expiration",
    "test_s3.test_lifecycle_noncur_expiration",
    "test_s3.test_lifecycle_deletemarker_expiration",
    "test_s3.test_lifecycle_multipart_expiration",
    "test_s3.test_encryption_sse_c_multipart_upload",
    "test_s3.test_encryption_sse_c_multipart_bad_download",
    "test_s3.test_sse_kms_multipart_upload",
    "test_s3.test_sse_kms_multipart_invalid_chunks_1",
    "test_s3.test_sse_kms_multipart_invalid_chunks_2",
    "test_s3.test_post_object_tags_anonymous_request",
    "test_s3.test_versioning_bucket_atomic_upload_return_version_id",
    "test_s3.test_versioning_bucket_multipart_upload_return_version_id",
    "test_s3.test_bucket_policy_put_obj_acl",
    "test_s3.test_bucket_policy_put_obj_grant",
    "test_s3.test_bucket_policy_put_obj_enc",
    "test_s3.test_bucket_policy_put_obj_request_obj_tag",
    "test_s3_website",
    "test_headers.test_object_acl_create_contentlength_none",
    "test_s3.test_bucket_list_return_data",
    "test_s3.test_multi_object_delete",
    "test_s3.test_object_raw_get",
    "test_s3.test_object_raw_get_bucket_gone",
    "test_s3.test_object_raw_get_object_gone",
    "test_s3.test_object_raw_get_bucket_acl",
    "test_s3.test_object_raw_get_object_acl",
    "test_s3.test_object_raw_authenticated",
    "test_s3.test_object_raw_response_headers",
    "test_s3.test_object_raw_authenticated_bucket_acl",
    "test_s3.test_object_raw_authenticated_object_acl",
    "test_s3.test_object_raw_authenticated_bucket_gone",
    "test_s3.test_object_raw_authenticated_object_gone",
    "test_s3.test_object_acl_default",
    "test_s3.test_object_acl_canned_during_create",
    "test_s3.test_object_acl_canned",
    "test_s3.test_object_acl_canned_publicreadwrite",
    "test_s3.test_object_acl_canned_authenticatedread",
    "test_s3.test_object_acl_xml",
    "test_s3.test_object_acl_xml_write",
    "test_s3.test_object_acl_xml_writeacp",
    "test_s3.test_object_acl_xml_read",
    "test_s3.test_object_acl_xml_readacp",
    "test_s3.test_bucket_acl_no_grants",
    "test_s3.test_object_header_acl_grants",
    "test_s3.test_object_set_valid_acl",
    "test_s3.test_object_giveaway",
    "test_s3.test_bucket_create_special_key_names",
    "test_s3.test_object_copy_zero_size",
    "test_s3.test_object_copy_same_bucket",
    "test_s3.test_object_copy_verify_contenttype",
    "test_s3.test_object_copy_to_itself_with_metadata",
    "test_s3.test_object_copy_diff_bucket",
    "test_s3.test_object_copy_not_owned_bucket",
    "test_s3.test_object_copy_not_owned_object_bucket",
    "test_s3.test_object_copy_canned_acl",
    "test_s3.test_object_copy_retaining_metadata",
    "test_s3.test_object_copy_replacing_metadata",
    "test_s3.test_multipart_copy_small",
    "test_s3.test_multipart_copy_invalid_range",
    "test_s3.test_multipart_copy_without_range",
    "test_s3.test_multipart_copy_special_names",
    "test_s3.test_multipart_copy_multiple_sizes",
    "test_s3.test_multipart_upload_overwrite_existing_object",
    "test_s3.test_atomic_multipart_upload_write",
    "test_s3.test_versioning_obj_create_read_remove",
    "test_s3.test_versioning_obj_create_read_remove_head",
    "test_s3.test_bucket_policy",
    "test_s3.test_bucket_policy_acl",
    "test_s3.test_bucket_policy_different_tenant",
    "test_s3.test_bucket_policy_another_bucket",
    "test_s3.test_bucket_policy_set_condition_operator_end_with_IfExists",
    "test_s3.test_bucket_policy_list_bucket_with_prefix",
    "test_s3.test_bucket_policy_list_bucket_with_maxkeys",
    "test_s3.test_bucket_policy_list_bucket_with_delimiter",
    "test_s3.test_bucket_policy_list_put_bucket_acl_canned_acl",
    "test_s3.test_bucket_policy_list_put_bucket_acl_grants",
    "test_s3.test_get_tags_acl_public",
    "test_s3.test_put_tags_acl_public",
    "test_s3.test_delete_tags_obj_public",
    "test_s3.test_bucket_policy_get_obj_existing_tag",
    "test_s3.test_bucket_policy_get_obj_tagging_existing_tag",
    "test_s3.test_bucket_policy_put_obj_tagging_existing_tag",
    "test_s3.test_bucket_policy_put_obj_copy_source",
    "test_s3.test_bucket_policy_put_obj_copy_source_meta",
    "test_s3.test_bucket_policy_get_obj_acl_existing_tag",
    "test_s3.test_bucket_list_delimiter_alt",
    "test_s3.test_bucket_list_delimiter_percentage",
    "test_s3.test_bucket_list_delimiter_whitespace",
    "test_s3.test_bucket_list_delimiter_dot",
    "test_s3.test_bucket_list_delimiter_unreadable",
    "test_s3.test_bucket_list_prefix_delimiter_alt",
    "test_s3.test_bucket_list_prefix_delimiter_delimiter_not_exist",
    "test_s3.test_bucket_list_prefix_delimiter_prefix_delimiter_not_exist",
    "test_headers.test_bucket_put_bad_canned_acl",
    "test_s3.test_post_object_invalid_date_format",
    "test_s3.test_post_object_no_key_specified",
    "test_s3.test_post_object_missing_signature",
    "test_s3.test_post_object_condition_is_case_sensitive",
    "test_s3.test_post_object_expires_is_case_sensitive",
    "test_s3.test_post_object_missing_expires_condition",
    "test_s3.test_post_object_missing_conditions_list",
    "test_s3.test_post_object_upload_size_limit_exceeded",
    "test_s3.test_post_object_missing_content_length_argument",
    "test_s3.test_post_object_invalid_content_length_argument",
    "test_s3.test_post_object_upload_size_below_minimum",
    "test_s3.test_post_object_empty_conditions",
    "test_s3.test_object_copy_to_itself",
    "test_s3.test_lifecycle_id_too_long",
    "test_s3.test_lifecycle_same_id",
    "test_s3.test_lifecycle_invalid_status",
    "test_s3.test_lifecycle_rules_conflicted",
    "test_s3.test_lifecycle_set_invalid_date",
    "test_s3.test_encryption_sse_c_multipart_invalid_chunks_1",
    "test_s3.test_encryption_sse_c_multipart_invalid_chunks_2",
    "test_s3.test_put_excess_tags",
    "test_s3.test_put_excess_key_tags",
    "test_s3.test_put_excess_val_tags",
    "test_s3.test_lifecycle_set_date",
    "test_s3.test_lifecycle_set_noncurrent",
    "test_s3.test_lifecycle_set_deletemarker",
    "test_s3.test_lifecycle_set_filter",
    "test_s3.test_lifecycle_set_empty_filter",
    "test_s3.test_lifecycle_set_multipart",
    "test_s3.test_get_obj_tagging",
    "test_s3.test_get_obj_head_tagging",
    "test_s3.test_put_max_tags",
    "test_s3.test_put_max_kvsize_tags",
    "test_s3.test_put_modify_tags",
    "test_s3.test_put_delete_tags",
    "test_s3.test_put_obj_with_tags",
    "test_s3.test_post_object_authenticated_request",
    "test_s3.test_post_object_authenticated_no_content_type",
    "test_s3.test_post_object_set_key_from_filename",
    "test_s3.test_post_object_ignored_header",
    "test_s3.test_post_object_case_insensitive_condition_fields",
    "test_s3.test_post_object_escaped_field_values",
    "test_s3.test_post_object_user_specified_header",
    "test_s3.test_encryption_sse_c_post_object_authenticated_request",
    "test_s3.test_sse_kms_post_object_authenticated_request",
    "test_s3.test_post_object_tags_authenticated_request",
    "test_s3.test_post_object_invalid_signature",
    "test_s3.test_post_object_invalid_access_key",
    "test_s3.test_post_object_missing_policy_condition",
    "test_s3.test_post_object_request_missing_policy_specified_field",
    "test_s3.test_post_object_expired_policy",
    "test_s3.test_post_object_invalid_request_field_value",
    "test_s3.test_object_copy_bucket_not_found",
    "test_s3.test_object_copy_key_not_found",
    "test_s3.test_abort_multipart_upload_not_found",
    "test_s3.test_set_cors",
    "test_headers.test_object_create_bad_authorization_empty",
    "test_headers.test_object_create_bad_authorization_none",
    "test_headers.test_bucket_create_bad_contentlength_empty",
    "test_headers.test_bucket_create_bad_authorization_empty",
    "test_headers.test_bucket_create_bad_authorization_none",
    "test_headers.test_object_create_bad_authorization_incorrect_aws2",
    "test_headers.test_object_create_bad_authorization_invalid_aws2",
    "test_s3.test_bucket_delete_nonowner",
    "test_s3.test_list_buckets_invalid_auth",
    "test_s3.test_list_buckets_bad_auth",
    "test_s3.test_encryption_sse_c_present",
    "test_s3.test_encryption_sse_c_other_key",
    "test_s3.test_encryption_sse_c_invalid_md5",
    "test_s3.test_encryption_sse_c_no_md5",
    "test_s3.test_encryption_sse_c_no_key",
    "test_s3.test_encryption_key_no_sse_c",
    "test_s3.test_sse_kms_no_key",
    "test_s3.test_sse_kms_not_declared",
    "test_s3.test_sse_kms_read_declare",
    "test_s3.test_sse_kms_method_head",
    "test_s3.test_encryption_sse_c_method_head",
    "test_s3.test_100_continue",
    "test_s3.test_object_set_get_metadata_none_to_empty"
    "test_s3.test_object_set_get_metadata_none_to_good",
    "test_s3.test_object_write_expires",
    "test_s3.test_object_write_cache_control",
    "test_s3.test_bucket_concurrent_set_canned_acl",
    "test_s3.test_object_set_get_metadata_overwrite_to_good",
    "test_s3.test_object_set_get_metadata_overwrite_to_empty",
    "test_s3.test_bucket_list_delimiter_prefix_ends_with_delimiter",
    "test_s3.test_bucket_list_objects_anonymous_fail",
    "test_headers.test_object_create_bad_date_invalid_aws2",
    "test_headers.test_object_create_bad_date_empty_aws2",
    "test_headers.test_object_create_bad_date_none_aws2",
    "test_headers.test_object_create_bad_date_before_today_aws2",
    "test_headers.test_object_create_bad_date_after_today_aws2",
    "test_headers.test_object_create_bad_date_before_epoch_aws2",
    "test_headers.test_object_create_bad_date_after_end_aws2",
    "test_headers.test_bucket_create_bad_authorization_invalid_aws2",
    "test_headers.test_bucket_create_bad_date_invalid_aws2",
    "test_headers.test_bucket_create_bad_date_empty_aws2",
    "test_headers.test_bucket_create_bad_date_none_aws2",
    "test_headers.test_bucket_create_bad_date_before_today_aws2",
    "test_headers.test_bucket_create_bad_date_after_today_aws2",
    "test_headers.test_bucket_create_bad_date_before_epoch_aws2",
    "test_s3.test_object_set_get_non_utf8_metadata",
    "test_s3.test_object_set_get_metadata_empty_to_unreadable_prefix",
    "test_s3.test_object_set_get_metadata_empty_to_unreadable_suffix",
    "test_s3.test_object_set_get_metadata_empty_to_unreadable_infix",
    "test_s3.test_object_set_get_metadata_overwrite_to_unreadable_prefix",
    "test_s3.test_object_set_get_metadata_overwrite_to_unreadable_suffix",
    "test_s3.test_object_set_get_metadata_overwrite_to_unreadable_infix",
    "test_s3.test_bucket_create_exists_nonowner",
    "test_headers.test_object_create_bad_authorization_unreadable",
    "test_headers.test_bucket_create_bad_authorization_unreadable",
    "test_headers.test_object_create_bad_date_unreadable_aws2",
    "test_headers.test_bucket_create_bad_date_unreadable_aws2",
    "test_s3.test_object_raw_put",
    "test_s3.test_object_raw_put_authenticated_expired",
    "test_s3.test_post_object_upload_larger_than_chunk",
    "test_s3.test_bucket_concurrent_set_canned_acl",
    "test_s3.test_bucket_list_delimiter_prefix_ends_with_delimiter",
    "test_s3.test_object_set_get_metadata_none_to_good",
    "test_s3.test_object_set_get_metadata_none_to_empty",
    "test_s3.test_object_set_get_unicode_metadata",
    "test_s3.test_bucket_create_naming_good_starts_alpha",
    "test_s3.test_bucket_create_naming_good_starts_digit",
    "test_s3.test_object_raw_get_x_amz_expires_out_range_zero",
    "test_s3.test_object_raw_get_x_amz_expires_out_max_range",
    "test_s3.test_object_raw_get_x_amz_expires_out_positive_range",
    "test_s3.test_object_anon_put",
    "test_s3.test_lifecycle_expiration_days0",
    "test_headers.test_object_create_bad_md5_unreadable",
    "test_headers.test_object_create_bad_contenttype_unreadable",
    "test_s3.test_bucket_list_prefix_delimiter_prefix_not_exist",
    "test_s3.test_lifecycle_expiration_header_put",
    "test_s3.test_lifecycle_expiration_header_head",
    "test_s3.test_list_buckets_anonymous",
    "test_s3.test_versioning_obj_read_not_exist_null",

    # These tests are disabled due to our tighter restrictions on bucket
    # naming
    "test_s3.test_bucket_create_naming_good_contains_period",
    "test_s3.test_bucket_create_naming_dns_dot_dot",
    "test_s3.test_bucket_create_naming_dns_dot_dash",
    "test_s3.test_bucket_create_naming_dns_dash_dot",

    # These tests are disabled because go's http server doesn't support custom
    # error responses for those triggered by these tests. If go adds support
    # for this, or if we put a reverse proxy in front of the s3gateway, these
    # tests could be run.
    "test_headers.test_object_create_bad_expect_unreadable",
    "test_headers.test_bucket_create_bad_expect_unreadable",
    "test_headers.test_object_create_bad_ua_unreadable_aws2",
    "test_headers.test_bucket_create_bad_ua_unreadable_aws2",
    "test_headers.test_bucket_create_bad_expect_mismatch",
    "test_headers.test_object_create_bad_expect_mismatch",

    # These tests are disabled for similar reasons. We could override errors
    # served back by go's built-in http facilities, but it would require
    # throwing out all of the convenient machinery that `http.ServeContent`
    # offers.
    "test_s3.test_ranged_request_invalid_range",
    "test_s3.test_ranged_request_empty_object",
    "test_s3.test_get_object_ifmatch_failed",
    "test_s3.test_get_object_ifunmodifiedsince_good",
    "test_s3.test_ranged_request_invalid_range",
    "test_s3.test_ranged_request_empty_object",
    "test_s3.test_put_object_ifmatch_failed",

    # These tests are disabled because go's http server automatically fixes
    # the error
    "test_headers.test_object_create_bad_contentlength_none",
    "test_headers.test_object_create_bad_contentlength_mismatch_above",
    "test_headers.test_object_create_bad_contentlength_empty",
    "test_headers.test_object_create_bad_contentlength_none",
    "test_headers.test_object_create_bad_contentlength_mismatch_below_aws2",

    # This test is disabled because the test's datetime validator does not
    # expect as much precision as what we return, which is arguably a bug in
    # the test itself
    "test_s3.test_get_object_ifmodifiedsince_failed",

    # This test is disabled because it treats `/foo/bar` as both a file and a
    # directory, which works in s3, but is not expected to in the s3gateway
    # TODO: server is returning 500's, and should give a better error for this
    "test_s3.test_bucket_list_delimiter_basic",

    # Ignored because it's testing non-standard extensions
    "test_s3.test_append_object",
    "test_s3.test_append_object_position_wrong",
    "test_s3.test_bucket_create_naming_bad_starts_nonalpha",
    "test_s3.test_atomic_write_bucket_gone",

    # These tests are disabled because the tests themselves have bugs:
    # - Seems to rely on an old boto bug
    "test_s3.test_object_create_unreadable",
    # - Doesn't preprend the test bucket name
    "test_s3.test_object_write_to_nonexist_bucket",
    # - Seems to be recreating the same bucket twice. Maybe it's expecting
    # that this bucket is being made in US-east-1, which has a special case in
    # AWS where recreating is a no-op.
    "test_s3.test_bucket_recreate_not_overriding",
    # - The correct error object is returned, but this test tries to use
    # attributes on the error object that don't exist
    "test_s3.test_bucket_create_exists",
    # - nosetests picks this up as a test function, even though it's intended
    # purpose is a utility function
    "test_s3.setup_lifecycle_expiration_test",
    # - This test checks that the returned etag matches an expected md5 value,
    # but the s3 spec says that etags should be treated as an opaque string,
    # and not tied to any specific algo
    "test_s3.test_object_write_check_etag",

    # these tests check that the returned etag matches an expected md5 value,
    # but the s3 spec says that etags should be treated as an opaque string,
    # and not tied to any specific algo
    "test_headers.test_object_create_bad_md5_none",
    "test_headers.test_object_create_bad_expect_empty",
    "test_headers.test_object_create_bad_expect_none",
    "test_headers.test_object_create_bad_contenttype_invalid",
    "test_headers.test_object_create_bad_contenttype_empty",
    "test_headers.test_object_create_bad_contenttype_none",
    "test_headers.test_object_create_date_and_amz_date",
    "test_headers.test_object_create_amz_date_and_no_date",
    "test_headers.test_object_create_bad_ua_empty_aws2",
    "test_headers.test_object_create_bad_ua_none_aws2",
    "test_s3.test_bucket_list_delimiter_empty",
    "test_s3.test_bucket_list_delimiter_none",
    "test_s3.test_bucket_list_delimiter_not_exist",
    "test_s3.test_bucket_list_prefix_alt",
    "test_s3.test_bucket_list_maxkeys_one",
    "test_s3.test_bucket_list_maxkeys_zero",
    "test_s3.test_bucket_list_maxkeys_none",
    "test_s3.test_bucket_list_maxkeys_invalid",
    "test_s3.test_bucket_list_maxkeys_unreadable",
    "test_s3.test_bucket_list_marker_none",
    "test_s3.test_bucket_list_marker_empty",
    "test_s3.test_bucket_list_marker_unreadable",
    "test_s3.test_bucket_list_marker_not_in_list",
    "test_s3.test_bucket_list_marker_after_list",
    "test_s3.test_bucket_list_marker_before_list",
    "test_s3.test_object_write_read_update_read_delete",
    "test_s3.test_object_metadata_replaced_on_put",
    "test_s3.test_object_write_file",
    "test_s3.test_get_object_ifmatch_good",
    "test_s3.test_get_object_ifnonematch_good",
    "test_s3.test_get_object_ifnonematch_failed",
    "test_s3.test_get_object_ifmodifiedsince_good",
    "test_s3.test_get_object_ifunmodifiedsince_failed",
    "test_s3.test_bucket_list_prefix_basic",
    "test_s3.test_bucket_list_prefix_empty",
    "test_s3.test_bucket_list_prefix_none",
    "test_s3.test_bucket_list_prefix_not_exist",
    "test_s3.test_bucket_list_prefix_unreadable",
    "test_s3.test_bucket_list_prefix_delimiter_basic",
    "test_s3.test_atomic_read_1mb",
    "test_s3.test_atomic_write_1mb",
    "test_s3.test_atomic_dual_write_1mb",
    "test_s3.test_encrypted_transfer_1MB",
    "test_s3.test_sse_kms_transfer_1MB",
    "test_s3.test_bucket_list_many",
    "test_s3.test_bucket_list_object_time",
    "test_s3.test_object_delete_key_bucket_gone",
    "test_s3.test_atomic_read_4mb",
    "test_s3.test_atomic_write_4mb",
    "test_s3.test_atomic_dual_write_4mb",
    "test_s3.test_atomic_read_8mb",
    "test_s3.test_atomic_write_8mb",
    "test_s3.test_atomic_dual_write_8mb",
    "test_s3.test_ranged_request_response_code",
    "test_s3.test_ranged_request_skip_leading_bytes_response_code",
    "test_s3.test_ranged_request_return_trailing_bytes_response_code",
    "test_s3.test_bucket_list_distinct",
    "test_s3.test_bucket_list_delimiter_prefix",
    "test_s3.test_encrypted_transfer_1b",
    "test_s3.test_sse_kms_transfer_1b",
    "test_s3.test_encrypted_transfer_1kb",
    "test_s3.test_sse_kms_transfer_1kb",
    "test_s3.test_encrypted_transfer_13b",
    "test_s3.test_sse_kms_transfer_13b",
    "test_s3.test_bucket_list_delimiter_prefix_underscore",
    "test_s3.test_bucket_delete_nonempty",
    "test_s3.test_object_head_zero_bytes",
    "test_s3.test_bucket_list_special_prefix",
    "test_s3.test_ranged_big_request_response_code",
    "test_s3.test_sse_kms_present",
    "test_s3.test_put_object_ifmatch_good",
    "test_s3.test_put_object_ifmatch_overwrite_existed_good",
    "test_s3.test_put_object_ifmatch_nonexisted_failed",
    "test_s3.test_put_object_ifnonmatch_good",
    "test_s3.test_put_object_ifnonmatch_failed",
    "test_s3.test_put_object_ifnonmatch_nonexisted_good",
    "test_s3.test_put_object_ifnonmatch_overwrite_existed_failed",
    "test_s3.test_atomic_conditional_write_1mb",
    "test_s3.test_atomic_dual_conditional_write_1mb",
    "test_s3.test_lifecycle_expiration_date",
    "test_s3.test_append_normal_object",
    "test_s3.test_bucket_head_extended",
]

# Ignores for boto3 functional tests
BLACKLISTED_BOTO3_FUNCTIONAL_TESTS = [
    # Most of the following tests are also ignored in boto2; see above
    # comments for why they are ignored.
    "test_headers.test_object_create_bad_contentlength_none",
    "test_headers.test_object_create_bad_authorization_empty",
    "test_headers.test_object_create_bad_authorization_none",
    "test_headers.test_bucket_create_bad_contentlength_empty",
    "test_headers.test_bucket_create_bad_authorization_empty",
    "test_headers.test_bucket_create_bad_authorization_none",
    "test_headers.test_object_create_bad_authorization_incorrect_aws2",
    "test_headers.test_object_create_bad_authorization_invalid_aws2",
    "test_headers.test_object_create_bad_date_invalid_aws2",
    "test_headers.test_object_create_bad_date_empty_aws2",
    "test_headers.test_object_create_bad_date_none_aws2",
    "test_headers.test_object_create_bad_date_before_today_aws2",
    "test_headers.test_object_create_bad_date_before_epoch_aws2",
    "test_headers.test_object_create_bad_date_after_end_aws2",
    "test_headers.test_bucket_create_bad_authorization_invalid_aws2",
    "test_headers.test_bucket_create_bad_date_invalid_aws2",
    "test_headers.test_bucket_create_bad_date_empty_aws2",
    "test_headers.test_bucket_create_bad_date_none_aws2",
    "test_headers.test_bucket_create_bad_date_before_today_aws2",
    "test_headers.test_bucket_create_bad_date_after_today_aws2",
    "test_headers.test_bucket_create_bad_date_before_epoch_aws2",
    "test_s3.test_bucket_list_objects_anonymous_fail",
    "test_s3.test_put_object_ifmatch_failed",
    "test_s3.test_object_raw_get_object_acl",
    "test_s3.test_object_anon_put",
    "test_s3.test_list_buckets_invalid_auth",
    "test_s3.test_list_buckets_bad_auth",
    "test_s3.test_encryption_sse_c_method_head",
    "test_s3.test_encryption_sse_c_present",
    "test_s3.test_encryption_sse_c_other_key",
    "test_s3.test_encryption_sse_c_invalid_md5",
    "test_s3.test_encryption_sse_c_no_md5",
    "test_s3.test_encryption_sse_c_no_key",
    "test_s3.test_encryption_key_no_sse_c",
    "test_s3.test_sse_kms_no_key",
    "test_s3.test_sse_kms_not_declared",
    "test_s3.test_sse_kms_read_declare",
    "test_headers.test_bucket_put_bad_canned_acl",
    "test_s3.test_post_object_invalid_date_format",
    "test_s3.test_post_object_no_key_specified",
    "test_s3.test_post_object_missing_signature",
    "test_s3.test_post_object_condition_is_case_sensitive",
    "test_s3.test_post_object_expires_is_case_sensitive",
    "test_s3.test_post_object_missing_expires_condition",
    "test_s3.test_post_object_missing_conditions_list",
    "test_s3.test_post_object_upload_size_limit_exceeded",
    "test_s3.test_post_object_missing_content_length_argument",
    "test_s3.test_post_object_invalid_content_length_argument",
    "test_s3.test_post_object_upload_size_below_minimum",
    "test_s3.test_post_object_empty_conditions",
    "test_s3.test_object_copy_to_itself",
    "test_s3.test_lifecycle_id_too_long",
    "test_s3.test_lifecycle_same_id",
    "test_s3.test_lifecycle_invalid_status",
    "test_s3.test_lifecycle_rules_conflicted",
    "test_s3.test_lifecycle_set_invalid_date",
    "test_s3.test_encryption_sse_c_multipart_invalid_chunks_1",
    "test_s3.test_encryption_sse_c_multipart_invalid_chunks_2",
    "test_s3.test_put_excess_tags",
    "test_s3.test_put_excess_key_tags",
    "test_s3.test_put_excess_val_tags",
    "test_s3.test_multipart_upload_empty",
    "test_s3.test_multipart_upload_small",
    "test_s3.test_multipart_copy_small",
    "test_s3.test_multipart_copy_invalid_range",
    "test_s3.test_multipart_copy_without_range",
    "test_s3.test_multipart_copy_special_names",
    "test_s3.test_multipart_upload_resend_part",
    "test_s3.test_multipart_upload_multiple_sizes",
    "test_s3.test_multipart_copy_multiple_sizes",
    "test_s3.test_multipart_upload_size_too_small",
    "test_s3.test_multipart_upload_contents",
    "test_s3.test_multipart_upload_overwrite_existing_object",
    "test_s3.test_abort_multipart_upload",
    "test_s3.test_list_multipart_upload",
    "test_s3.test_multipart_upload_missing_part",
    "test_s3.test_multipart_upload_incorrect_etag",
    "test_s3.test_atomic_multipart_upload_write",
    "test_s3.test_multipart_resend_first_finishes_last",
    "test_s3.test_encryption_sse_c_multipart_bad_download",
    "test_s3.test_sse_kms_multipart_upload",
    "test_s3.test_sse_kms_multipart_invalid_chunks_1",
    "test_s3.test_sse_kms_multipart_invalid_chunks_2",
    "test_s3.test_bucket_list_return_data_versioning",
    "test_s3.test_object_copy_versioned_bucket",
    "test_s3.test_object_copy_versioning_multipart_upload",
    "test_s3.test_versioning_obj_create_read_remove",
    "test_s3.test_versioning_obj_create_read_remove_head",
    "test_s3.test_versioning_obj_suspend_versions",
    "test_s3.test_versioning_obj_create_versions_remove_all",
    "test_s3.test_versioning_obj_create_versions_remove_special_names",
    "test_s3.test_versioning_obj_create_overwrite_multipart",
    "test_s3.test_versioning_obj_list_marker",
    "test_s3.test_versioning_copy_obj_version",
    "test_s3.test_versioning_multi_object_delete",
    "test_s3.test_versioning_multi_object_delete_with_marker",
    "test_s3.test_versioning_multi_object_delete_with_marker_create",
    "test_s3.test_versioned_object_acl",
    "test_s3.test_versioned_object_acl_no_version_specified",
    "test_s3.test_versioned_concurrent_object_create_concurrent_remove",
    "test_s3.test_versioned_concurrent_object_create_and_remove",
    "test_s3.test_versioning_bucket_atomic_upload_return_version_id",
    "test_s3.test_versioning_bucket_multipart_upload_return_version_id",
    "test_s3.test_bucket_policy",
    "test_s3.test_bucket_policy_different_tenant",
    "test_s3.test_bucket_policy_another_bucket",
    "test_s3.test_bucket_policy_set_condition_operator_end_with_IfExists",
    "test_s3.test_get_tags_acl_public",
    "test_s3.test_put_tags_acl_public",
    "test_s3.test_delete_tags_obj_public",
    "test_s3.test_bucket_policy_get_obj_existing_tag",
    "test_s3.test_bucket_policy_get_obj_tagging_existing_tag",
    "test_s3.test_bucket_policy_put_obj_tagging_existing_tag",
    "test_s3.test_bucket_policy_put_obj_copy_source",
    "test_s3.test_bucket_policy_put_obj_copy_source_meta",
    "test_s3.test_bucket_policy_put_obj_acl",
    "test_s3.test_bucket_policy_put_obj_grant",
    "test_s3.test_bucket_policy_put_obj_enc",
    "test_s3.test_bucket_policy_put_obj_request_obj_tag",
    "test_s3.test_bucket_policy_get_obj_acl_existing_tag",
    "test_s3.test_post_object_anonymous_request",
    "test_s3.test_post_object_authenticated_request",
    "test_s3.test_post_object_authenticated_no_content_type",
    "test_s3.test_post_object_set_invalid_success_code",
    "test_s3.test_post_object_set_key_from_filename",
    "test_s3.test_post_object_ignored_header",
    "test_s3.test_post_object_case_insensitive_condition_fields",
    "test_s3.test_post_object_escaped_field_values",
    "test_s3.test_post_object_user_specified_header",
    "test_s3.test_encryption_sse_c_post_object_authenticated_request",
    "test_s3.test_sse_kms_post_object_authenticated_request",
    "test_s3.test_post_object_tags_anonymous_request",
    "test_s3.test_post_object_tags_authenticated_request",
    "test_s3.test_bucket_list_objects_anonymous",
    "test_s3.test_bucket_acl_canned_private_to_private",
    "test_s3.test_access_bucket_private_object_private",
    "test_s3.test_access_bucket_private_object_publicread",
    "test_s3.test_access_bucket_private_object_publicreadwrite",
    "test_s3.test_access_bucket_publicread_object_private",
    "test_s3.test_access_bucket_publicread_object_publicread",
    "test_s3.test_access_bucket_publicread_object_publicreadwrite",
    "test_s3.test_access_bucket_publicreadwrite_object_private",
    "test_s3.test_access_bucket_publicreadwrite_object_publicread",
    "test_s3.test_access_bucket_publicreadwrite_object_publicreadwrite",
    "test_s3.test_bucket_policy_acl",
    "test_s3.test_bucket_acl_default",
    "test_s3.test_bucket_acl_canned",
    "test_s3.test_bucket_acl_canned_publicreadwrite",
    "test_s3.test_bucket_acl_canned_authenticatedread",
    "test_s3.test_object_acl_canned_bucketownerread",
    "test_s3.test_object_acl_canned_bucketownerfullcontrol",
    "test_s3.test_bucket_acl_grant_nonexist_user",
    "test_s3.test_bucket_acl_no_grants",
    "test_s3.test_bucket_acl_grant_email_notexist",
    "test_s3.test_bucket_acl_revoke_all",
    "test_s3.test_lifecycle_set",
    "test_s3.test_lifecycle_get",
    "test_s3.test_lifecycle_get_no_id",
    "test_s3.test_lifecycle_set_date",
    "test_s3.test_lifecycle_expiration_days0",
    "test_s3.test_lifecycle_set_noncurrent",
    "test_s3.test_lifecycle_set_deletemarker",
    "test_s3.test_lifecycle_set_filter",
    "test_s3.test_lifecycle_set_empty_filter",
    "test_s3.test_lifecycle_set_multipart",
    "test_s3.test_bucket_list_delimiter_alt",
    "test_s3.test_bucket_list_delimiter_percentage",
    "test_s3.test_bucket_list_delimiter_whitespace",
    "test_s3.test_bucket_list_delimiter_dot",
    "test_s3.test_bucket_list_delimiter_unreadable",
    "test_s3.test_bucket_list_prefix_delimiter_alt",
    "test_s3.test_bucket_list_prefix_delimiter_prefix_not_exist",
    "test_s3.test_bucket_list_prefix_delimiter_delimiter_not_exist",
    "test_s3.test_bucket_list_prefix_delimiter_prefix_delimiter_not_exist",
    "test_s3.test_bucket_list_return_data",
    "test_s3.test_object_acl_default",
    "test_s3.test_object_acl_canned_during_create",
    "test_s3.test_object_acl_canned",
    "test_s3.test_object_acl_canned_publicreadwrite",
    "test_s3.test_object_acl_canned_authenticatedread",
    "test_s3.test_object_acl_full_control_verify_attributes",
    "test_s3.test_object_copy_not_owned_object_bucket",
    "test_s3.test_object_copy_zero_size",
    "test_s3.test_object_copy_same_bucket",
    "test_s3.test_object_copy_verify_contenttype",
    "test_s3.test_object_copy_to_itself_with_metadata",
    "test_s3.test_object_copy_diff_bucket",
    "test_s3.test_object_copy_canned_acl",
    "test_s3.test_object_copy_retaining_metadata",
    "test_s3.test_object_copy_replacing_metadata",
    "test_s3.test_post_object_authenticated_request_bad_access_key",
    "test_s3.test_post_object_invalid_signature",
    "test_s3.test_post_object_invalid_access_key",
    "test_s3.test_post_object_missing_policy_condition",
    "test_s3.test_post_object_request_missing_policy_specified_field",
    "test_s3.test_post_object_expired_policy",
    "test_s3.test_post_object_invalid_request_field_value",
    "test_s3.test_object_copy_not_owned_bucket",
    "test_s3.test_object_set_get_non_utf8_metadata",
    "test_s3.test_object_set_get_metadata_empty_to_unreadable_prefix",
    "test_s3.test_object_set_get_metadata_empty_to_unreadable_suffix",
    "test_s3.test_object_set_get_metadata_empty_to_unreadable_infix",
    "test_s3.test_object_set_get_metadata_overwrite_to_unreadable_prefix",
    "test_s3.test_object_set_get_metadata_overwrite_to_unreadable_suffix",
    "test_s3.test_object_set_get_metadata_overwrite_to_unreadable_infix",
    "test_s3.test_get_obj_tagging",
    "test_s3.test_get_obj_head_tagging",
    "test_s3.test_put_max_tags",
    "test_s3.test_put_max_kvsize_tags",
    "test_s3.test_put_modify_tags",
    "test_s3.test_put_delete_tags",
    "test_s3.test_multipart_copy_versioned",
    "test_s3.test_versioning_bucket_create_suspend",
    "test_s3.test_versioning_obj_plain_null_version_removal",
    "test_s3.test_versioning_obj_plain_null_version_overwrite",
    "test_s3.test_versioning_obj_plain_null_version_overwrite_suspended",
    "test_s3.test_abort_multipart_upload_not_found",
    "test_s3.test_set_cors",
    "test_s3.test_cors_origin_response",
    "test_s3.test_cors_origin_wildcard",
    "test_s3.test_cors_header_option",
    "test_s3.test_object_set_get_metadata_none_to_good",
    "test_s3.test_object_set_get_metadata_none_to_empty",
    "test_s3.test_object_set_get_metadata_overwrite_to_empty",
    "test_s3.test_object_set_get_unicode_metadata",
    "test_s3.test_object_raw_get_x_amz_expires_out_range_zero",
    "test_s3.test_object_raw_get_x_amz_expires_out_max_range",
    "test_s3.test_object_raw_get_x_amz_expires_out_positive_range",
    "test_s3.test_object_raw_put_authenticated_expired",
    "test_s3.test_bucket_create_naming_good_starts_alpha",
    "test_s3.test_bucket_create_naming_good_starts_digit",
    "test_headers.test_object_acl_create_contentlength_none",
    "test_s3.test_bucket_create_special_key_names",
    "test_s3.test_get_object_ifmatch_failed",
    "test_s3.test_get_object_ifunmodifiedsince_good",
    "test_s3.test_ranged_request_invalid_range",
    "test_s3.test_ranged_request_empty_object",
    "test_headers.test_object_create_bad_expect_mismatch",
    "test_headers.test_object_create_bad_contentlength_mismatch_above",
    "test_headers.test_bucket_create_bad_expect_mismatch",
    "test_s3.test_bucket_list_delimiter_basic",
    "test_s3.test_bucket_list_delimiter_prefix_ends_with_delimiter",
    "test_s3.test_multi_object_delete",
    "test_s3.test_object_write_cache_control",
    "test_s3.test_object_write_expires",
    "test_s3.test_post_object_upload_larger_than_chunk",
    "test_s3.test_bucket_create_exists",
    "test_s3.test_logging_toggle",
    "test_s3.test_bucket_create_naming_good_contains_period",
    "test_s3.test_bucket_recreate_not_overriding",
    "test_s3.test_sse_kms_method_head",
    "test_s3.test_put_obj_with_tags",
    "test_headers.test_object_create_bad_contentlength_mismatch_below_aws2",
    "test_s3.test_bucket_concurrent_set_canned_acl",
    "test_s3.test_object_write_to_nonexist_bucket",
    "test_s3.test_object_write_check_etag",
    "test_s3.test_post_object_set_success_code",
    "test_s3.test_post_object_success_redirect_action",
    "test_s3.test_object_raw_response_headers",
    "test_s3.test_bucket_create_exists_nonowner",
    "test_s3.test_100_continue",
    "test_s3.test_lifecycle_expiration_header_put",
    "test_s3.test_lifecycle_expiration_header_head",
    "test_s3.setup_lifecycle_expiration_test",
    "test_s3.test_bucket_acl_canned_during_create",
    "test_s3.test_bucket_acl_grant_userid_fullcontrol",
    "test_s3.test_bucket_acl_grant_userid_read",
    "test_s3.test_bucket_acl_grant_userid_readacp",
    "test_s3.test_bucket_acl_grant_userid_write",
    "test_s3.test_bucket_acl_grant_userid_writeacp",
    "test_s3.test_bucket_header_acl_grants",
    "test_s3.test_bucket_acl_grant_email",
    "test_s3.test_object_acl",
    "test_s3.test_object_acl_write",
    "test_s3.test_object_acl_writeacp",
    "test_s3.test_object_acl_read",
    "test_s3.test_object_acl_readacp",
    "test_s3.test_object_header_acl_grants",
    "test_s3.test_bucket_create_naming_dns_dot_dot",
    "test_s3.test_bucket_create_naming_dns_dot_dash",
    "test_s3.test_bucket_create_naming_dns_dash_dot",
    "test_s3.test_multipart_upload",
    "test_s3.test_lifecycle_multipart_expiration",
    "test_s3.test_encryption_sse_c_multipart_upload",
    "test_s3.test_lifecycle_expiration",
    "test_s3.test_lifecycle_expiration_date",
    "test_s3.test_lifecycle_noncur_expiration",
    "test_s3.test_lifecycle_deletemarker_expiration",
    "test_s3.test_object_acl_full_control_verify_owner",
    "test_s3.test_put_object_ifmatch_nonexisted_failed",
    "test_s3.test_list_buckets_anonymous",

    # Ignored because it's testing non-standard extensions
    "test_s3.test_bucket_list_unordered",
    "test_s3.test_bucket_create_naming_bad_starts_nonalpha",
    "test_s3.test_bucket_head_extended",
    "test_s3.test_put_object_ifnonmatch_failed",
    "test_s3.test_put_object_ifnonmatch_overwrite_existed_failed",
    "test_s3.test_atomic_dual_conditional_write_1mb",
]

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
                _, skipped_str, _, errored_str, _, failed_str = match.groups()
                skipped = int(skipped_str) if skipped_str is not None else 0
                errored = int(errored_str) if errored_str is not None else 0
                failed = int(failed_str) if failed_str is not None else 0

    if ran != 0:
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
                    failing_test = match.groups()[1]
            else:
                if not any(line.startswith(p) for p in TRACEBACK_PREFIXES):
                    causes[line].append(failing_test)
                    failing_test = None

    causes = sorted(causes.items(), key=lambda i: len(i[1]), reverse=True)
    for (cause_name, failing_tests) in causes:
        if len(cause_name) > 160:
            print("{} [...]:".format(cause_name[:160]))
        else:
            print("{}:".format(cause_name))
        
        for failing_test in failing_tests:
            print("- {}".format(failing_test))

    return 0

def main():
    parser = argparse.ArgumentParser(description="Runs a conformance test suite for PFS' s3gateway.")
    parser.add_argument("--no-run", default=False, action="store_true", help="Disables a test run, and just prints failure data from the last test run")
    parser.add_argument("--test", default="", help="Run a specific test")
    args = parser.parse_args()

    if args.no_run:
        sys.exit(print_failures())

    proc = subprocess.run("yes | pachctl delete-all", shell=True)
    if proc.returncode != 0:
        raise Exception("bad exit code: {}".format(proc.returncode))

    if args.test:
        print("Running test {}".format(args.test))

        # In some places, nose and its plugins expect tests to be
        # specified as testmodule.testname, but here, it's expected to be
        # testmodule:testname. This replaces the last . with a : so that
        # the testmodule.testname format can be used everywhere, including
        # here.
        if "." in test and not ":" in test:
            test = ":".join(args.test.rsplit(".", 1))
        else:
            test = args.test

        run_nosetests(test)
    else:
        print("Running all tests")

        # This uses the `nose-exclude` plugin to exclude tests for
        # unsupported features. Note that `nosetest` does have a built-in
        # way of excluding tests, but it only seems to match on top-level
        # modules, rather than on specific tests.
        blacklisted_tests = []
        for test in BLACKLISTED_BOTO2_FUNCTIONAL_TESTS:
            blacklisted_tests.append("s3tests.functional.{}".format(test))
        for test in BLACKLISTED_BOTO3_FUNCTIONAL_TESTS:
            blacklisted_tests.append("s3tests_boto3.functional.{}".format(test))
        extra_env = {
            "NOSE_EXCLUDE_TESTS": ";".join(blacklisted_tests)
        }

        filepath = os.path.join(RUNS_ROOT, datetime.datetime.now().strftime("%Y-%m-%d-%H-%M-%S.txt"))
        with open(filepath, "w") as f:
            run_nosetests(env=extra_env, stderr=f)

        sys.exit(print_failures())
                
if __name__ == "__main__":
    main()
