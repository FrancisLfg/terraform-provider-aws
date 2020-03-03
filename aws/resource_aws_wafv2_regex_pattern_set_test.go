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

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafv2RegexPatternSet(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":          testAccAWSWafv2RegexPatternSet_basic,
		//"changePatterns": testAccAWSWafv2RegexPatternSet_changePatterns,
		//"noPatterns":     testAccAWSWafv2RegexPatternSet_noPatterns,
		//"disappears":     testAccAWSWafv2RegexPatternSet_disappears,
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
	resourceName := "aws_waf_v2_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafv2RegexPatternSetConfig(patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),

					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "description", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regular_expression_list", wafv2.ParameterExceptionFieldRegexPatternReferenceStatement),
					//resource.TestCheckResourceAttr(resourceName, "RegexString.#", "2"),
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
			},
		},
	})
}

func testAccAWSWafv2RegexPatternSet_changePatterns(t *testing.T) {
	var before, after wafv2.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafv2RegexPatternSetConfig(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.2848565413", "one"),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.3351840846", "two"),
				),
			},
			{
				Config: testAccAWSWafv2RegexPatternSetConfig_changePatterns(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.3351840846", "two"),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.2929247714", "three"),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.1294846542", "four"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSWafv2RegexPatternSet_noPatterns(t *testing.T) {
	var patternSet wafv2.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafv2RegexPatternSetConfig_noPatterns(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &patternSet),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//func testAccAWSWafv2RegexPatternSet_disappears(t *testing.T) {
//	var v wafv2.RegexPatternSet
//	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
//	resourceName := "aws_waf_regex_pattern_set.test"
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
//		Providers:    testAccProviders,
//		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccAWSWafv2RegexPatternSetConfig(patternSetName),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
//					testAccCheckAWSWafv2RegexPatternSetDisappears(&v),
//				),
//				ExpectNonEmptyPlan: true,
//			},
//		},
//	})
//}
//
//func testAccCheckAWSWafv2RegexPatternSetDisappears(set *wafv2.RegexPatternSet) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
//
//		wr := newWafRetryer(conn)
//		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
//			req := &wafv2.UpdateRegexPatternSetInput{
//				ChangeToken:       token,
//				RegexPatternSetId: set.RegexPatternSetId,
//			}
//
//			for _, pattern := range set.RegexPatternStrings {
//				update := &wafv2.RegexPatternSetUpdate{
//					Action:             aws.String("DELETE"),
//					RegexPatternString: pattern,
//				}
//				req.Updates = append(req.Updates, update)
//			}
//
//			return conn.UpdateRegexPatternSet(req)
//		})
//		if err != nil {
//			return fmt.Errorf("Failed updating WAF Regex Pattern Set: %s", err)
//		}
//
//		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
//			opts := &wafv2.DeleteRegexPatternSetInput{
//				ChangeToken:       token,
//				RegexPatternSetId: set.RegexPatternSetId,
//			}
//			return conn.DeleteRegexPatternSet(opts)
//		})
//		if err != nil {
//			return fmt.Errorf("Failed deleting WAF Regex Pattern Set: %s", err)
//		}
//
//		return nil
//	}
//}

func testAccCheckAWSWafv2RegexPatternSetExists(n string, v *wafv2.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regex Pattern Set ID is set")
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

		return fmt.Errorf("WAF Regex Pattern Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafv2RegexPatternSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_regex_pattern_set" {
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
				return fmt.Errorf("WAF Regex Pattern Set %s still exists", rs.Primary.ID)
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

func testAccAWSWafv2RegexPatternSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, name)
}

func testAccAWSWafv2RegexPatternSetConfig_changePatterns(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["two", "three", "four"]
}
`, name)
}

func testAccAWSWafv2RegexPatternSetConfig_noPatterns(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name = "%s"
}
`, name)
}
