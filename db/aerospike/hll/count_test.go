//go:build integration
// +build integration

package hll

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"testing"

	aero "github.com/InsideGallery/core/db/aerospike"
	"github.com/InsideGallery/core/db/aerospike/testfixtures"
	"github.com/InsideGallery/core/mathutils"
	as "github.com/aerospike/aerospike-client-go/v7"
)

var hllPolicy = as.NewHLLPolicy(as.HLLWriteFlagsDefault | as.HLLWriteFlagsNoFail)

var client *as.Client

func setup() ([]testfixtures.AerospikeFixture, error) {
	// todo: change client
	if _, ok := os.LookupEnv("AEROSPIKE_HOST"); !ok {
		return nil, fmt.Errorf("AEROSPIKE_HOST env var is not set")
	}

	var asErr error
	client, asErr = as.NewClient(os.Getenv("AEROSPIKE_HOST"), 3000)
	if asErr != nil {
		return nil, fmt.Errorf("failed to create aerospike client: %w", asErr)
	}

	log.Printf("Aerospike client created!")

	fixtures, err := loadTestFixtures(client, "../../../fixtures/aerospike/hll_count_test_data.json")
	if err != nil {
		return fixtures, fmt.Errorf("load test fixtures err: %w", err)
	}
	log.Printf("Test fixtures loaded!")

	return fixtures, nil
}

func teardown(fixtures []testfixtures.AerospikeFixture) error {
	if err := testfixtures.CleanupAerospikeFixtures(client, fixtures); err != nil {
		return fmt.Errorf("failed to cleanup test fixtures: %s", err)
	}

	if client != nil {
		client.Close()
	}

	return nil
}

func TestMain(m *testing.M) {
	// Call setup() before running the tests
	fixtures, err := setup()
	if err != nil {
		log.Printf("failed to setup tests err: %s", err.Error())
		return
	}

	// Run the tests
	exitCode := m.Run()

	// Call teardown() after running the tests
	if err := teardown(fixtures); err != nil {
		log.Fatalf("failed to teardown tests err: %s", err)
	}

	// Exit with the proper exit code
	os.Exit(exitCode)
}

func loadTestFixtures(client *as.Client, filename string) ([]testfixtures.AerospikeFixture, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read test fixtures logfile: %w", err)
	}

	fixtures, err := testfixtures.LoadAerospikeFixtures(client, data)
	if err != nil {
		return nil, err
	}
	return fixtures, nil
}

func TestCount_WithOneField(t *testing.T) {
	const (
		attr      = "account_email"
		namespace = "transactions"
		set       = "account"

		accountEmail = "sergay@test.com"
	)

	errorTolerance := 0.02

	// t.Logf("error tolerance: %f", errorTolerance)

	values := []as.Value{
		as.StringValue("test1@gmail.com"),
		as.StringValue("test2@gmail.com"),
		as.StringValue("test3@gmail.com"),
		as.StringValue("test4@gmail.com"),
	}

	var ops []*as.Operation
	ops = append(ops,
		// add operation to null value of hll bin
		as.HLLInitOp(hllPolicy, aero.HLLBin, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
		as.HLLAddOp(hllPolicy, aero.HLLBin, values, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
	)

	key, asErr := as.NewKey(namespace, set, strings.Join([]string{attr, accountEmail}, ":"))
	if asErr != nil {
		t.Fatalf("failed to create key: %s", asErr)
	}

	_, asErr = client.Operate(nil, key, ops...)
	if asErr != nil {
		t.Fatalf("failed to operate: %s", asErr)
	}

	count, err := CountIntersection(client, namespace, set, []string{strings.Join([]string{attr, accountEmail}, ":")})
	if err != nil {
		t.Fatalf("failed to count: %s", err)
	}

	cntError := math.Abs(float64(count-int64(len(values)))) / float64(len(values))
	if cntError > errorTolerance {
		t.Fatalf("Error: count is %d, actual count is %d, relative error is %.3f, tolerance is %.3f",
			count, len(values), cntError, errorTolerance)
	}
}

func TestCount_WithMultipleFields(t *testing.T) {
	fixtures, err := loadTestFixtures(client, "../../../fixtures/aerospike/hll_count_test_data.json")
	if err != nil {
		t.Fatalf("failed to load test fixtures: %s", err)
	}
	defer func() {
		if err := testfixtures.CleanupAerospikeFixtures(client, fixtures); err != nil {
			t.Fatalf("failed to cleanup test fixtures: %s", err)
		}
	}()

	const (
		errorTolerance = 0.03

		accountEmailValue       = "sergay@test.com"
		testTransactionCurrency = "UAH"
		testRequestID           = "53"
	)

	type Field struct {
		name     string
		keyValue string

		values []as.Value
	}

	tt := []struct {
		name      string
		namespace string
		set       string
		fields    []Field
		wantCount int64
	}{
		{
			name:      "2 fields with 3 intersecting values",
			namespace: "transactions",
			set:       "standard",
			fields: []Field{
				{
					name:     "account_email",
					keyValue: "account_email:" + accountEmailValue,
					values: []as.Value{
						as.StringValue("test1"),
						as.StringValue("test2"),
						as.StringValue("test"),
						as.StringValue("test013"),
						as.StringValue("test02"),
						as.StringValue("test099"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
					},
				},
				{
					name:     "transaction_currency",
					keyValue: "transaction_currency:" + testTransactionCurrency,
					values: []as.Value{
						as.StringValue("test1"),
						as.StringValue("test2"),
						as.StringValue("test"),
						as.StringValue("test77"),
						as.StringValue("test902"),
						as.StringValue("test9123"),
						as.StringValue("test1241"),
					},
				},
			},
			wantCount: 3,
		},
		{
			name:      "3 fields with 10 intersecting values",
			namespace: "transactions",
			set:       "standard",
			fields: []Field{
				{
					name:     "account_email",
					keyValue: "account_email:" + accountEmailValue,
					values: []as.Value{
						as.StringValue("test1"),
						as.StringValue("test2"),
						as.StringValue("test3"),
						as.StringValue("test4"),
						as.StringValue("test5"),
						as.StringValue("test6"),
						as.StringValue("test7"),
						as.StringValue("test8"),
						as.StringValue("test9"),
						as.StringValue("test10"),
						as.StringValue("test10"),
						as.StringValue("test10"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
					},
				},
				{
					name:     "transaction_currency",
					keyValue: "transaction_currency:" + testTransactionCurrency,
					values: []as.Value{
						as.StringValue("test1"),
						as.StringValue("test2"),
						as.StringValue("test3"),
						as.StringValue("test4"),
						as.StringValue("test5"),
						as.StringValue("test6"),
						as.StringValue("test7"),
						as.StringValue("test8"),
						as.StringValue("test9"),
						as.StringValue("test10"),
					},
				},
				{
					name:     "request_id",
					keyValue: "request_id:" + testRequestID,
					values: []as.Value{
						as.StringValue("test1"),
						as.StringValue(""),
						as.StringValue(""),
						as.StringValue("23"),
						as.StringValue("test2"),
						as.StringValue("test3"),
						as.StringValue("test4"),
						as.StringValue("test5"),
						as.StringValue("test6"),
						as.StringValue("test7"),
						as.StringValue("test8"),
						as.StringValue("test9"),
						as.StringValue("test10"),
					},
				},
			},
			wantCount: 10,
		},
		{
			name:      "3 fields with 0 intersecting values",
			namespace: "transactions",
			set:       "standard",
			fields: []Field{
				{
					name:     "account_email",
					keyValue: "account_email:" + accountEmailValue,
					values: []as.Value{
						as.StringValue("test099"),
						as.StringValue("test00092"),
						as.StringValue("test00092"),
					},
				},
				{
					name:     "transaction_currency",
					keyValue: "transaction_currency:" + testTransactionCurrency,
					values: []as.Value{
						as.StringValue("test902"),
						as.StringValue("test9123"),
					},
				},
				{
					name:     "request_id",
					keyValue: "request_id:" + testRequestID,
					values: []as.Value{
						as.StringValue("test9"),
						as.StringValue("test10"),
					},
				},
			},
			wantCount: 0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var ops []*as.Operation
			var by []string
			for _, field := range tc.fields {
				ops = append(ops,
					// add operation to null value of hll bin
					as.HLLInitOp(hllPolicy, aero.HLLBin, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
					as.HLLAddOp(hllPolicy, aero.HLLBin, field.values, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
				)
				key, err := as.NewKey(tc.namespace, tc.set, field.keyValue)
				if err != nil {
					t.Fatalf("failed to create key: %s", err)
				}

				_, err = client.Operate(nil, key, ops...)
				if err != nil {
					t.Fatalf("failed to operate: %s", err)
				}
				by = append(by, field.keyValue)
			}

			count, err := CountIntersection(client, tc.namespace, tc.set, by)
			if err != nil {
				t.Fatalf("failed to count: %s", err)
			}

			// do not check error tolerance if wantCount is 0 divide by 0 exception
			if tc.wantCount == 0 && count == 0 {
				return
			}
			cntError := math.Abs(float64(count-tc.wantCount)) / float64(tc.wantCount)
			// t.Logf("count: %d, want: %d, error tolerance: %f", count, tc.wantCount, errorTolerance)

			if cntError > errorTolerance {
				t.Fatalf("Error: count is %d, actual count is %d, relative error is %.3f, tolerance is %.3f",
					count, tc.wantCount, cntError, errorTolerance)
			}
		})
	}
}

func BenchmarkCountIntersection_5ValuesLen20(b *testing.B) {
	if !client.IsConnected() {
		b.Fatal("aerospike client is not connected")
	}

	const (
		namespace = "transactions"
		set       = "account"
		attrName  = "account_email"
		keyValue  = "account_email:hll_benchmark"
	)

	var values []as.Value
	for i := 0; i < 5; i++ {
		values = append(values, as.StringValue(mathutils.RandomDigitString(20)))
	}

	var ops []*as.Operation
	ops = append(ops,
		// add operation to null value of hll bin
		as.HLLInitOp(hllPolicy, aero.HLLBin, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
		as.HLLAddOp(hllPolicy, aero.HLLBin, values, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
	)
	key, err := as.NewKey(namespace, set, keyValue)
	if err != nil {
		b.Fatalf("failed to create key: %s", err)
	}

	_, err = client.Operate(nil, key, ops...)
	if err != nil {
		b.Fatalf("failed to operate: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CountIntersection(client, namespace, set, []string{keyValue})
		if err != nil {
			b.Fatalf("failed to count: %s", err)
		}
	}
}

func BenchmarkCountIntersection_5ValuesLen3(b *testing.B) {
	if !client.IsConnected() {
		b.Fatal("aerospike client is not connected")
	}

	const (
		namespace = "transactions"
		set       = "account"
		attrName  = "account_email"
		keyValue  = "account_email:hll_benchmark"
	)

	var values []as.Value
	for i := 0; i < 5; i++ {
		values = append(values, as.StringValue(mathutils.RandomDigitString(3)))
	}

	var ops []*as.Operation
	ops = append(ops,
		// add operation to null value of hll bin
		as.HLLInitOp(hllPolicy, aero.HLLBin, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
		as.HLLAddOp(hllPolicy, aero.HLLBin, values, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
	)
	key, err := as.NewKey(namespace, set, keyValue)
	if err != nil {
		b.Fatalf("failed to create key: %s", err)
	}

	_, err = client.Operate(nil, key, ops...)
	if err != nil {
		b.Fatalf("failed to operate: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CountIntersection(client, namespace, set, []string{keyValue})
		if err != nil {
			b.Fatalf("failed to count: %s", err)
		}
	}
}

func BenchmarkCountIntersection_MillionValuesLen20(b *testing.B) {
	if !client.IsConnected() {
		b.Fatal("aerospike client is not connected")
	}

	const (
		namespace = "transactions"
		set       = "account"
		attrName  = "account_email"
		keyValue  = "account_email:hll_benchmark"
	)

	var values []as.Value
	for i := 0; i < 1000000; i++ {
		values = append(values, as.StringValue(mathutils.RandomDigitString(20)))
	}

	var ops []*as.Operation
	ops = append(ops,
		// add operation to null value of hll bin
		as.HLLInitOp(hllPolicy, aero.HLLBin, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
		as.HLLAddOp(hllPolicy, aero.HLLBin, values, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
	)
	key, err := as.NewKey(namespace, set, keyValue)
	if err != nil {
		b.Fatalf("failed to create key: %s", err)
	}

	_, err = client.Operate(nil, key, ops...)
	if err != nil {
		b.Fatalf("failed to operate: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CountIntersection(client, namespace, set, []string{keyValue})
		if err != nil {
			b.Fatalf("failed to count: %s", err)
		}
	}
}

func BenchmarkCountIntersection_MillionValuesLen3(b *testing.B) {
	if !client.IsConnected() {
		b.Fatal("aerospike client is not connected")
	}

	const (
		namespace = "transactions"
		set       = "account"
		attrName  = "account_email"
		keyValue  = "account_email:hll_benchmark"
	)

	var values []as.Value
	for i := 0; i < 1000000; i++ {
		values = append(values, as.StringValue(mathutils.RandomDigitString(3)))
	}

	var ops []*as.Operation
	ops = append(ops,
		// add operation to null value of hll bin
		as.HLLInitOp(hllPolicy, aero.HLLBin, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
		as.HLLAddOp(hllPolicy, aero.HLLBin, values, aero.MaxIndexBits, aero.MaxAllowedMinhashBits),
	)
	key, err := as.NewKey(namespace, set, keyValue)
	if err != nil {
		b.Fatalf("failed to create key: %s", err)
	}

	_, err = client.Operate(nil, key, ops...)
	if err != nil {
		b.Fatalf("failed to operate: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CountIntersection(client, namespace, set, []string{keyValue})
		if err != nil {
			b.Fatalf("failed to count: %s", err)
		}
	}
}
