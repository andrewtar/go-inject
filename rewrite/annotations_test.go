package rewrite

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/monnoroch/go-inject"
)

const testValue = 17

type testAnnotation1 struct{}
type testAnnotation2 struct{}
type testAnnotation3 struct{}
type testAnnotation4 struct{}

type RewriteAnnotationsTests struct {
	suite.Suite
}

type testModuleWithProviders struct {
	providers []inject.Provider
}

func (self testModuleWithProviders) Providers() ([]inject.Provider, error) {
	return self.providers, nil
}

func (self *RewriteAnnotationsTests) TestNoArgumentsProvider() {
	providers := self.getProviders(
		testModuleWithProviders{[]inject.Provider{inject.NewProvider(
			func() (int, testAnnotation1) {
				return testValue, testAnnotation1{}
			},
		)}},
		map[inject.Annotation]inject.Annotation{},
	)
	self.Equal(1, len(providers))
	provider := providers[0]
	self.False(provider.IsCached())

	value, annotation := self.call(provider, []reflect.Value{})

	self.Equal(testValue, value)

	_, ok := annotation.(testAnnotation1)
	self.True(ok)
}

func (self *RewriteAnnotationsTests) TestProviderWithArguments() {
	providers := self.getProviders(
		testModuleWithProviders{[]inject.Provider{inject.NewProvider(
			func(value int, _ testAnnotation2) (int, testAnnotation1) {
				return value + 1, testAnnotation1{}
			},
		)}},
		map[inject.Annotation]inject.Annotation{},
	)
	self.Equal(1, len(providers))
	provider := providers[0]
	self.False(provider.IsCached())

	value, annotation := self.call(provider, []reflect.Value{
		reflect.ValueOf(int(testValue)),
		reflect.ValueOf(testAnnotation2{}),
	})

	self.Equal(testValue+1, value)

	_, ok := annotation.(testAnnotation1)
	self.True(ok)
}

func (self *RewriteAnnotationsTests) TestReplaceNoArgumentsProvider() {
	providers := self.getProviders(
		testModuleWithProviders{[]inject.Provider{inject.NewProvider(
			func() (int, testAnnotation1) {
				return testValue, testAnnotation1{}
			},
		)}},
		map[inject.Annotation]inject.Annotation{
			testAnnotation1{}: testAnnotation3{},
			testAnnotation2{}: testAnnotation4{},
		},
	)
	self.Equal(1, len(providers))
	provider := providers[0]
	self.False(provider.IsCached())

	value, annotation := self.call(provider, []reflect.Value{})

	self.Equal(testValue, value)

	_, ok := annotation.(testAnnotation3)
	self.True(ok)
}

func (self *RewriteAnnotationsTests) TestReplaceProviderWithArguments() {
	providers := self.getProviders(
		testModuleWithProviders{[]inject.Provider{inject.NewProvider(
			func(value int, _ testAnnotation2) (int, testAnnotation1) {
				return value + 1, testAnnotation1{}
			},
		)}},
		map[inject.Annotation]inject.Annotation{
			testAnnotation1{}: testAnnotation3{},
			testAnnotation2{}: testAnnotation4{},
		},
	)
	self.Equal(1, len(providers))
	provider := providers[0]
	self.False(provider.IsCached())

	value, annotation := self.call(provider, []reflect.Value{
		reflect.ValueOf(int(testValue)),
		reflect.ValueOf(testAnnotation4{}),
	})

	self.Equal(testValue+1, value)

	_, ok := annotation.(testAnnotation3)
	self.True(ok)
}

func (self *RewriteAnnotationsTests) TestCachedProvider() {
	providers := self.getProviders(
		testModuleWithProviders{[]inject.Provider{inject.NewProvider(func() (int, testAnnotation1) {
			return 0, testAnnotation1{}
		}).Cached(true)}},
		map[inject.Annotation]inject.Annotation{},
	)
	self.Equal(1, len(providers))
	self.True(providers[0].IsCached())
}

type testErrorModule struct {
	err error
}

func (self testErrorModule) Providers() ([]inject.Provider, error) {
	return nil, self.err
}

func (self *RewriteAnnotationsTests) TestProvidersError() {
	testError := errors.New("test error")
	_, err := SubstituteAnnotations(testErrorModule{testError}, map[inject.Annotation]inject.Annotation{}).Providers()
	self.Equal(testError, err)
}

func (self *RewriteAnnotationsTests) TestInvalidProvider() {
	_, err := SubstituteAnnotations(
		testModuleWithProviders{[]inject.Provider{inject.NewProvider(0)}},
		map[inject.Annotation]inject.Annotation{},
	).Providers()
	self.Contains(err.Error(), "invalid provider")
}

func (self *RewriteAnnotationsTests) getProviders(
	module inject.Module,
	annotationsToRewrite map[inject.Annotation]inject.Annotation,
) []inject.Provider {
	providers, err := SubstituteAnnotations(module, annotationsToRewrite).Providers()
	self.Require().Nil(err)
	return providers
}

func (self *RewriteAnnotationsTests) call(
	provider inject.Provider,
	arguments []reflect.Value,
) (interface{}, interface{}) {
	outputs := provider.Function().Call(arguments)
	self.Equal(2, len(outputs))
	return outputs[0].Interface(), outputs[1].Interface()
}

func TestRewriteAnnotations(t *testing.T) {
	suite.Run(t, new(RewriteAnnotationsTests))
}