package rules

import (
	"errors"
	"fmt"
	"testing"

	C "github.com/MerlinKodo/clash-rev/constant"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRule(t *testing.T) {
	type testCase struct {
		tp            C.RuleConfig
		payload       string
		target        string
		params        []string
		expectedRule  C.Rule
		expectedError error
	}

	policy := "DIRECT"

	testCases := []testCase{
		{
			tp:           C.RuleConfigDomain,
			payload:      "example.com",
			target:       policy,
			expectedRule: NewDomain("example.com", policy),
		},
		{
			tp:           C.RuleConfigDomainSuffix,
			payload:      "example.com",
			target:       policy,
			expectedRule: NewDomainSuffix("example.com", policy),
		},
		{
			tp:           C.RuleConfigDomainKeyword,
			payload:      "example.com",
			target:       policy,
			expectedRule: NewDomainKeyword("example.com", policy),
		},
		{
			tp:      C.RuleConfigGeoIP,
			payload: "CN",
			target:  policy, params: []string{noResolve},
			expectedRule: NewGEOIP("CN", policy, true),
		},
		{
			tp:           C.RuleConfigIPCIDR,
			payload:      "127.0.0.0/8",
			target:       policy,
			expectedRule: lo.Must(NewIPCIDR("127.0.0.0/8", policy, WithIPCIDRNoResolve(false))),
		},
		{
			tp:      C.RuleConfigIPCIDR,
			payload: "127.0.0.0/8",
			target:  policy, params: []string{noResolve},
			expectedRule: lo.Must(NewIPCIDR("127.0.0.0/8", policy, WithIPCIDRNoResolve(true))),
		},
		{
			tp:           C.RuleConfigIPCIDR6,
			payload:      "2001:db8::/32",
			target:       policy,
			expectedRule: lo.Must(NewIPCIDR("2001:db8::/32", policy, WithIPCIDRNoResolve(false))),
		},
		{
			tp:      C.RuleConfigIPCIDR6,
			payload: "2001:db8::/32",
			target:  policy, params: []string{noResolve},
			expectedRule: lo.Must(NewIPCIDR("2001:db8::/32", policy, WithIPCIDRNoResolve(true))),
		},
		{
			tp:           C.RuleConfigSrcIPCIDR,
			payload:      "192.168.1.201/32",
			target:       policy,
			expectedRule: lo.Must(NewIPCIDR("192.168.1.201/32", policy, WithIPCIDRSourceIP(true), WithIPCIDRNoResolve(true))),
		},
		{
			tp:           C.RuleConfigSrcPort,
			payload:      "80",
			target:       policy,
			expectedRule: lo.Must(NewPort("80", policy, PortTypeSrc)),
		},
		{
			tp:           C.RuleConfigDstPort,
			payload:      "80",
			target:       policy,
			expectedRule: lo.Must(NewPort("80", policy, PortTypeDest)),
		},
		{
			tp:           C.RuleConfigInboundPort,
			payload:      "80",
			target:       policy,
			expectedRule: lo.Must(NewPort("80", policy, PortTypeInbound)),
		},
		{
			tp:           C.RuleConfigProcessName,
			payload:      "example.exe",
			target:       policy,
			expectedRule: lo.Must(NewProcess("example.exe", policy, true)),
		},
		{
			tp:           C.RuleConfigProcessPath,
			payload:      "C:\\Program Files\\example.exe",
			target:       policy,
			expectedRule: lo.Must(NewProcess("C:\\Program Files\\example.exe", policy, false)),
		},
		{
			tp:           C.RuleConfigProcessPath,
			payload:      "/opt/example/example",
			target:       policy,
			expectedRule: lo.Must(NewProcess("/opt/example/example", policy, false)),
		},
		{
			tp:      C.RuleConfigIPSet,
			payload: "example",
			target:  policy,
			// unit test runs on Linux machine and NewIPSet(...) won't be available
			expectedError: errors.New("operation not permitted"),
		},
		{
			tp:      C.RuleConfigIPSet,
			payload: "example",
			target:  policy, params: []string{noResolve},
			// unit test runs on Linux machine and NewIPSet(...) won't be available
			expectedError: errors.New("operation not permitted"),
		},
		{
			tp:           C.RuleConfigMatch,
			payload:      "example",
			target:       policy,
			expectedRule: NewMatch(policy),
		},
		{
			tp:            C.RuleConfigRuleSet,
			payload:       "example",
			target:        policy,
			expectedError: fmt.Errorf("unsupported rule type %s", C.RuleConfigRuleSet),
		},
		{
			tp:            C.RuleConfigScript,
			payload:       "example",
			target:        policy,
			expectedError: fmt.Errorf("unsupported rule type %s", C.RuleConfigScript),
		},
		{
			tp:            "UNKNOWN",
			payload:       "example",
			target:        policy,
			expectedError: errors.New("unsupported rule type UNKNOWN"),
		},
		{
			tp:            "ABCD",
			payload:       "example",
			target:        policy,
			expectedError: errors.New("unsupported rule type ABCD"),
		},
	}

	for _, tc := range testCases {
		_, err := ParseRule(string(tc.tp), tc.payload, tc.target, tc.params)
		if tc.expectedError != nil {
			require.Error(t, err)
			assert.EqualError(t, err, tc.expectedError.Error())
		} else {
			require.NoError(t, err)
		}
	}
}
