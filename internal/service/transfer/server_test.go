// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(transfer.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Invalid server type: PUBLIC",
		"InvalidServiceName: The Vpc Endpoint Service",
	)
}

func testAccServer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	iamRoleResourceName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "endpoint", "server.transfer", regexp.MustCompile(`s-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "function", ""),
					resource.TestCheckNoResourceAttr(resourceName, "host_key"),
					resource.TestCheckResourceAttrSet(resourceName, "host_key_fingerprint"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "invocation_role", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_role", ""),
					resource.TestCheckResourceAttr(resourceName, "post_authentication_login_banner", ""),
					resource.TestCheckResourceAttr(resourceName, "pre_authentication_login_banner", ""),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.as2_transports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.passive_ip", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.set_stat_option", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.tls_session_resumption_mode", "ENFORCED"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "SFTP"),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "domain", "S3"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "endpoint", "server.transfer", regexp.MustCompile(`s-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "function", ""),
					resource.TestCheckNoResourceAttr(resourceName, "host_key"),
					resource.TestCheckResourceAttrSet(resourceName, "host_key_fingerprint"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "invocation_role", ""),
					resource.TestCheckResourceAttrPair(resourceName, "logging_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "SFTP"),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
				),
			},
		},
	})
}

func testAccServer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceServer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServerConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccServer_domain(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_domain(rName, "EFS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "domain", "EFS"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_securityPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_securityPolicy(rName, "TransferSecurityPolicy-2020-06"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2020-06"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_securityPolicy(rName, "TransferSecurityPolicy-2018-11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
				),
			},
			{
				Config: testAccServerConfig_securityPolicy(rName, "TransferSecurityPolicy-2022-03"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2022-03"),
				),
			},
			{
				Config: testAccServerConfig_securityPolicy(rName, "TransferSecurityPolicy-2023-05"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2023-05"),
				),
			},
		},
	})
}

func testAccServer_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_vpcUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccServerConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_vpcAddressAllocationIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	eip1ResourceName := "aws_eip.test.0"
	eip2ResourceName := "aws_eip.test.1"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpcAddressAllocationIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_vpcAddressAllocationIdsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccServerConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test2"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_vpcSecurityGroupIdsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_vpcAddressAllocationIds_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	eip1ResourceName := "aws_eip.test.0"
	eip2ResourceName := "aws_eip.test.1"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test2"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpcAddressAllocationIdsSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_vpcAddressAllocationIdsSecurityGroupIdsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_updateEndpointType_publicToVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				Config: testAccServerConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_publicToVPC_addressAllocationIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				Config: testAccServerConfig_vpcAddressAllocationIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcEndpointToVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				Config: testAccServerConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcEndpointToVPC_addressAllocationIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
				),
			},
			{
				Config: testAccServerConfig_vpcAddressAllocationIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcEndpointToVPC_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				Config: testAccServerConfig_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcToPublic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_structuredLogDestinations(t *testing.T) {
	ctx := acctest.Context(t)
	var s transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	cloudwatchLogGroupName := "aws_cloudwatch_log_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_structuredLogDestinations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &s),
					// resource.TestCheckTypeSetElemAttr(resourceName, "structured_logging_destinations.*", *s.StructuredLogDestinations[0]),
					resource.ComposeTestCheckFunc(func(s *terraform.State) error {
						cwResource, ok := s.RootModule().Resources[cloudwatchLogGroupName]
						if !ok {
							return fmt.Errorf("resource not found: %s", cloudwatchLogGroupName)
						}
						cwARN, ok := cwResource.Primary.Attributes["arn"]
						if !ok {
							return errors.New("cloudwatch group arn missing")
						}
						expectedSLD := fmt.Sprintf("%s:*", cwARN)
						transferServerResource, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						slds, ok := transferServerResource.Primary.Attributes["structured_log_destinations"]
						if !ok {
							return errors.New("transfer server structured logging destinations missing")
						}
						if expectedSLD != slds {
							return fmt.Errorf("'%s' != '%s'", expectedSLD, slds)
						}
						return nil
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_structuredLogDestinationsUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &s),
					// resource.TestCheckTypeSetElemAttr(resourceName, "structured_logging_destinations.*", *s.StructuredLogDestinations[0]),
					// resource.TestCheckTypeSetElemAttr(resourceName, "structured_logging_destinations.*", fmt.Sprintf("\"${%s.arn}:*\"", cloudwatchLogGroupName)),
					resource.ComposeTestCheckFunc(func(s *terraform.State) error {
						cwResource, ok := s.RootModule().Resources[cloudwatchLogGroupName]
						if !ok {
							return fmt.Errorf("resource not found: %s", cloudwatchLogGroupName)
						}
						cwARN, ok := cwResource.Primary.Attributes["arn"]
						if !ok {
							return errors.New("cloudwatch group arn missing")
						}
						expectedSLD := fmt.Sprintf("%s:*", cwARN)
						transferServerResource, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						slds, ok := transferServerResource.Primary.Attributes["structured_logging_destinations"]
						if !ok {
							return errors.New("transfer server structured logging destinations missing")
						}
						if expectedSLD != slds {
							return fmt.Errorf("'%s' != '%s'", expectedSLD, slds)
						}
						return nil
					}),
				),
			},
		},
	})
}

func testAccServer_protocols(t *testing.T) {
	ctx := acctest.Context(t)
	var s transfer.DescribedServer
	var ca acmpca.CertificateAuthority
	resourceName := "aws_transfer_server.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.RandomDomainName()
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_protocols(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &s),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTP"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccServerConfig_rootCA(rootDomain),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(ctx, acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(ctx, &ca),
				),
			},
			{
				Config: testAccServerConfig_protocolsUpdate(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &s),
					resource.TestCheckResourceAttrPair(resourceName, "certificate", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTP"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTPS"),
				),
			},
			{
				Config: testAccServerConfig_protocolsUpdate(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(ctx, &ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServer_protocolDetails(t *testing.T) {
	ctx := acctest.Context(t)
	var s transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_protocolDetails("AUTO", "DEFAULT", "ENFORCED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &s),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.as2_transports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.passive_ip", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.set_stat_option", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.tls_session_resumption_mode", "ENFORCED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_protocolDetails("8.8.8.8", "ENABLE_NO_OP", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &s),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.as2_transports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.passive_ip", "8.8.8.8"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.set_stat_option", "ENABLE_NO_OP"),
					resource.TestCheckResourceAttr(resourceName, "protocol_details.0.tls_session_resumption_mode", "DISABLED"),
				),
			},
		},
	})
}

func testAccServer_apiGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_apiGatewayIdentityProviderType(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_role", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_apiGateway_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_apiGatewayIdentityProviderType(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_role", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_directoryService(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_directoryServiceIdentityProviderType(rName, domain, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "AWS_DIRECTORY_SERVICE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var s transfer.DescribedServer
	var u transfer.DescribedUser
	var k transfer.SshPublicKey
	resourceName := "aws_transfer_server.test"
	userResourceName := "aws_transfer_user.test"
	sshKeyResourceName := "aws_transfer_ssh_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_forceDestroy(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &s),
					testAccCheckUserExists(ctx, userResourceName, &u),
					testAccCheckSSHKeyExists(ctx, sshKeyResourceName, &k),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccServer_hostKey(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	hostKey := "test-fixtures/transfer-ssh-rsa-key"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_hostKey(rName, hostKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "host_key_fingerprint", "SHA256:Z2pW9sPKDD/T34tVfCoolsRcECNTlekgaKvDn9t+9sg="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccServer_vpcEndpointID(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_vpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_endpoint_id", vpcEndpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.vpc_id", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccServer_lambdaFunction(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_lambdaFunctionIdentityProviderType(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "function", "aws_lambda_function.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "AWS_LAMBDA"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_authenticationLoginBanners(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_displayBanners(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "post_authentication_login_banner", "This system is for the use of authorized users only - post"),
					resource.TestCheckResourceAttr(resourceName, "pre_authentication_login_banner", "This system is for the use of authorized users only - pre"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_workflowDetails(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig_workflow(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.0.on_partial_upload.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_partial_upload.0.execution_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_partial_upload.0.workflow_id", "aws_transfer_workflow.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.0.on_upload.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.execution_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.workflow_id", "aws_transfer_workflow.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerConfig_workflowUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.0.on_partial_upload.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_partial_upload.0.execution_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_partial_upload.0.workflow_id", "aws_transfer_workflow.test2", "id"),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.0.on_upload.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.execution_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.workflow_id", "aws_transfer_workflow.test2", "id"),
				),
			},
			{
				Config: testAccServerConfig_workflowRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.#", "0"),
				),
			},
		},
	})
}

func testAccCheckServerExists(ctx context.Context, n string, v *transfer.DescribedServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Server ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		output, err := tftransfer.FindServerByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckServerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_server" {
				continue
			}

			_, err := tftransfer.FindServerByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Server %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccServerConfig_vpcBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_subnet" "test2" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true

  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_eip" "test" {
  count = 2

  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_loggingRoleBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoCloudWatchLogs",
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}
`, rName)
}

func testAccServerConfig_apiGatewayBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = %[1]q
  stage_description = %[1]q

  variables = {
    "a" = "2"
  }
}
`, rName)
}

const testAccServerConfig_basic = `
resource "aws_transfer_server" "test" {}
`

func testAccServerConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccServerConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccServerConfig_displayBanners() string {
	return `
resource "aws_transfer_server" "test" {
  pre_authentication_login_banner  = "This system is for the use of authorized users only - pre"
  post_authentication_login_banner = "This system is for the use of authorized users only - post"
}
`
}

func testAccServerConfig_domain(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  domain = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, domain)
}

func testAccServerConfig_securityPolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  security_policy_name = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, policy)
}

func testAccServerConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_loggingRoleBase(rName), `
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.test.arn

  # No tags.
}
`)
}

func testAccServerConfig_apiGatewayIdentityProviderType(rName string, forceDestroy bool) string {
	return acctest.ConfigCompose(testAccServerConfig_apiGatewayBase(rName), testAccServerConfig_loggingRoleBase(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn

  force_destroy = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, forceDestroy))
}

func testAccServerConfig_directoryServiceIdentityProviderType(rName, domain string, forceDestroy bool) string {
	return acctest.ConfigCompose(
		testAccServerConfig_vpcBase(rName),
		testAccServerConfig_loggingRoleBase(rName),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"

  vpc_settings {
    vpc_id = aws_vpc.test.id

    subnet_ids = [
      aws_subnet.test.id,
      aws_subnet.test2.id
    ]
  }
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "AWS_DIRECTORY_SERVICE"
  directory_id           = aws_directory_service_directory.test.id
  logging_role           = aws_iam_role.test.arn

  force_destroy = %[3]t

  tags = {
    Name = %[1]q
  }
}
`, rName, domain, forceDestroy))
}

func testAccServerConfig_forceDestroy(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  force_destroy = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoS3",
    "Effect": "Allow",
    "Action": [
      "s3:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = %[1]q
  role      = aws_iam_role.test.arn
}

resource "aws_transfer_ssh_key" "test" {
  server_id = aws_transfer_server.test.id
  user_name = aws_transfer_user.test.user_name
  body      = "%[2]s"
}
`, rName, publicKey)
}

func testAccServerConfig_vpcEndpoint(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "transfer.server"
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  vpc_endpoint_type = "Interface"
  service_name      = data.aws_vpc_endpoint_service.test.service_name

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_transfer_server" "test" {
  endpoint_type = "VPC_ENDPOINT"

  endpoint_details {
    vpc_endpoint_id = aws_vpc_endpoint.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpc(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    vpc_id = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpcUpdate(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpcAddressAllocationIDs(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[0].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpcAddressAllocationIdsUpdate(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[1].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpcAddressAllocationIdsSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[0].id]
    security_group_ids     = [aws_security_group.test.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpcAddressAllocationIdsSecurityGroupIdsUpdate(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[1].id]
    security_group_ids     = [aws_security_group.test2.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    security_group_ids = [aws_security_group.test.id]
    vpc_id             = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_vpcSecurityGroupIdsUpdate(rName string) string {
	return acctest.ConfigCompose(testAccServerConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    security_group_ids = [aws_security_group.test2.id]
    vpc_id             = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_hostKey(rName, hostKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  host_key = file(%[2]q)

  tags = {
    Name = %[1]q
  }
}
`, rName, hostKey)
}

func testAccServerConfig_structuredLogDestinations(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
		resource "aws_cloudwatch_log_group" "test" {
			name_prefix = "transfer_test_"
		  }
		  
		  data "aws_iam_policy_document" "test" {
			statement {
			  effect = "Allow"
		  
			  principals {
				type        = "Service"
				identifiers = ["transfer.amazonaws.com"]
			  }
		  
			  actions = ["sts:AssumeRole"]
			}
		  }
		  
		  resource "aws_iam_role" "test" {
			name_prefix         = "iam_for_transfer_"
			assume_role_policy  = data.aws_iam_policy_document.test.json
			managed_policy_arns = ["arn:aws:iam::aws:policy/service-role/AWSTransferLoggingAccess"]
		  }
		  
		  resource "aws_transfer_server" "test" {
			endpoint_type = "PUBLIC"
			logging_role  = aws_iam_role.test.arn
			protocols     = ["SFTP"]
			structured_log_destinations = [
			  "${aws_cloudwatch_log_group.test.arn}:*"
			]
			tags = {
			  Name = %[1]q
			}
		  }
		`, rName),
	)
}

func testAccServerConfig_structuredLogDestinationsUpdate() string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
		resource "aws_cloudwatch_log_group" "test" {
			name_prefix = "transfer_test_"
		  }
		  
		  data "aws_iam_policy_document" "test" {
			statement {
			  effect = "Allow"
		  
			  principals {
				type        = "Service"
				identifiers = ["transfer.amazonaws.com"]
			  }
		  
			  actions = ["sts:AssumeRole"]
			}
		  }
		  
		  resource "aws_iam_role" "test" {
			name_prefix         = "iam_for_transfer_"
			assume_role_policy  = data.aws_iam_policy_document.test.json
			managed_policy_arns = ["arn:aws:iam::aws:policy/service-role/AWSTransferLoggingAccess"]
		  }
		  
		  resource "aws_transfer_server" "test" {
			endpoint_type = "PUBLIC"
			logging_role  = aws_iam_role.test.arn
			protocols     = ["SFTP"]
			structured_log_destinations = [
			  "${aws_cloudwatch_log_group.test.arn}:*"
			]
		  }
		`),
	)
}

func testAccServerConfig_protocols(rName string) string {
	return acctest.ConfigCompose(
		testAccServerConfig_vpcBase(rName),
		testAccServerConfig_apiGatewayBase(rName),
		testAccServerConfig_loggingRoleBase(rName),
		fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
  protocols              = ["FTP"]

  endpoint_type = "VPC"
  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerConfig_protocolDetails(passive_ip, set_stat_option, tls_session_resumption_mode string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  protocol_details {
    passive_ip                  = %[1]q
    set_stat_option             = %[2]q
    tls_session_resumption_mode = %[3]q
  }
}
`, passive_ip, set_stat_option, tls_session_resumption_mode)
}

func testAccServerConfig_rootCA(domain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}
`, domain)
}

func testAccServerConfig_protocolsUpdate(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(
		testAccServerConfig_vpcBase(rName),
		testAccServerConfig_apiGatewayBase(rName),
		testAccServerConfig_loggingRoleBase(rName),
		testAccServerConfig_rootCA(rootDomain),
		fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[2]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
  protocols              = ["FTP", "FTPS"]
  certificate            = aws_acm_certificate.test.arn

  endpoint_type = "VPC"
  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, domain))
}

func testAccServerConfig_lambdaFunctionIdentityProviderType(rName string, forceDestroy bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccServerConfig_loggingRoleBase(rName+"-logging"),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs14.x"
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "AWS_LAMBDA"
  function               = aws_lambda_function.test.arn
  logging_role           = aws_iam_role.test.arn

  force_destroy = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, forceDestroy))
}

func testAccServerConfig_workflow(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = "test"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}

resource "aws_transfer_server" "test" {
  workflow_details {
    on_upload {
      execution_role = aws_iam_role.test.arn
      workflow_id    = aws_transfer_workflow.test.id
    }
    on_partial_upload {
      execution_role = aws_iam_role.test.arn
      workflow_id    = aws_transfer_workflow.test.id
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccServerConfig_workflowUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = "test"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}

resource "aws_transfer_workflow" "test2" {
  steps {
    delete_step_details {
      name                 = "test"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}

resource "aws_transfer_server" "test" {
  workflow_details {
    on_upload {
      execution_role = aws_iam_role.test.arn
      workflow_id    = aws_transfer_workflow.test2.id
    }
    on_partial_upload {
      execution_role = aws_iam_role.test.arn
      workflow_id    = aws_transfer_workflow.test2.id
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccServerConfig_workflowRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_transfer_workflow" "test2" {
  steps {
    delete_step_details {
      name                 = "test"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}

resource "aws_transfer_server" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
