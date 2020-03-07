package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Serialized acceptance tests due to WAFV2 account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafv2RegexPatternSet(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":          testAccAWSWafv2RegexPatternSet_basic,
		"changePatterns": testAccAWSWafv2RegexPatternSet_changePatterns,
		"minimal":        TestAccAwsWafv2RegexPatternSet_minimal,
		"force_new":      TestAccAwsWafv2RegexPatternSet_changeNameForceNew,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAWSWafv2RegexPatternSet_basic(t *testing.T) {
	var v wafv2.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafv2RegexPatternSetConfig(patternSetName, wafv2.ScopeRegional),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),

					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "description", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regular_expression_list.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSWafv2RegexPatternSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccAWSWafv2RegexPatternSet_changePatterns(t *testing.T) {
	var before, after wafv2.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafv2RegexPatternSetConfig(patternSetName, wafv2.ScopeRegional),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regular_expression_list.#", "2"),
				),
			},
			{
				Config: testAccAWSWafv2RegexPatternSetConfigChangePatterns(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regular_expression_list.#", "3"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RegexPatternSet_minimal(t *testing.T) {
	var v wafv2.RegexPatternSet
	regexPatternSetName := fmt.Sprintf("regex-pattern-set-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RegexPatternSetConfigMinimal(regexPatternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", regexPatternSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression_list.#", "1"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RegexPatternSet_changeNameForceNew(t *testing.T) {
	var before, after wafv2.RegexPatternSet
	regexPatternSetName := fmt.Sprintf("regex-pattern-set-%s", acctest.RandString(5))
	regexPatternSetNewName := fmt.Sprintf("regex-pattern-set-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafv2RegexPatternSetConfig(regexPatternSetName, wafv2.ScopeRegional),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", regexPatternSetName),
					resource.TestCheckResourceAttr(resourceName, "description", regexPatternSetName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression_list.#", "2"),
				),
			},
			{
				Config: testAccAWSWafv2RegexPatternSetConfig(regexPatternSetNewName, wafv2.ScopeRegional),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", regexPatternSetNewName),
					resource.TestCheckResourceAttr(resourceName, "description", regexPatternSetNewName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression_list.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAWSWafv2RegexPatternSetExists(n string, v *wafv2.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFV2 Regex Pattern Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRegexPatternSet(&wafv2.GetRegexPatternSetInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if *resp.RegexPatternSet.Id == rs.Primary.ID {
			*v = *resp.RegexPatternSet
			return nil
		}

		return fmt.Errorf("WAFV2 Regex Pattern Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafv2RegexPatternSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_regex_pattern_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRegexPatternSet(&wafv2.GetRegexPatternSetInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err == nil {
			if *resp.RegexPatternSet.Id == rs.Primary.ID {
				return fmt.Errorf("WAFV2 Regex Pattern Set %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the Regex Pattern Set is already destroyed
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafv2RegexPatternSetConfig(name string, scope string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name                    = "%s"
  description             = "%s"
  scope                   = "%s"
  regular_expression_list = ["^foobar$","^example$"]
  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
`, name, name, scope)
}

func testAccAWSWafv2RegexPatternSetConfigChangePatterns(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name                    = "%s"
  description             = "%s"
  scope                   = "REGIONAL"
  regular_expression_list = ["^foobar$","^example$", "another"]
  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
`, name, name)
}

func testAccAwsWafv2RegexPatternSetConfigMinimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name                    = "%s"
  scope                   = "REGIONAL"
  regular_expression_list = ["^foobar$"]
}
`, name)
}

func testAccAWSWafv2RegexPatternSetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
