module github.com/mosuke5/sample-controller-operatorsdk

go 1.13

require (
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/motemen/go-quickfix v0.0.0-20200118031250-2a6e54e79a50 // indirect
	github.com/motemen/gore v0.5.0 // indirect
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/operator-framework/operator-sdk-samples/go/memcached-operator v0.0.0-20200428142309-9bb42cdb16f5
	github.com/peterh/liner v1.2.0 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/tools v0.0.0-20200504022951-6b6965ac5dd1 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
)
