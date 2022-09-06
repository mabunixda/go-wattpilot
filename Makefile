all: preprocess fmt build

preprocess:
	curl -OL https://raw.githubusercontent.com/joscha82/wattpilot/da08c3fb387b06497e007bef1ff88f0112a080ea/src/wattpilot/ressources/wattpilot.yaml
	rm -f wattpilot_mapping_gen.go
	echo "package api" >> wattpilot_mapping_gen.go
	echo "var propertyMap = map[string]string {" >> wattpilot_mapping_gen.go
	cat wattpilot.yaml | yq e '.properties[] | select( .key != "pvopt_deltaA" and .key != "pvopt_deltaP")  |  ( "\"" + .alias  + "\":\"" + .key + "\",")' >> wattpilot_mapping_gen.go
	echo "}" >> wattpilot_mapping_gen.go

fmt:
	go fmt ./...

build:
	CURPWD=$(PWD)
	cd shell
	go build
	cd $(CURPWD)
