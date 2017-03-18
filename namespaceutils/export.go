package namespaceutils

import (
	"encoding/json"
	"reflect"
	"strings"

	squallmodels "github.com/aporeto-inc/gaia/squallmodels/current/golang"

	"github.com/aporeto-inc/elemental"
	"github.com/aporeto-inc/manipulate"
)

func ContentOfNamespace(manipulator manipulate.Manipulator, namespace string) (elemental.IdentifiablesList, error) {

	identifiablesChannel := make(chan elemental.IdentifiablesList)
	errorsChannel := make(chan error)
	identifiables := elemental.IdentifiablesList{}

	mctx := manipulate.NewContext()
	mctx.Recursive = true
	mctx.Namespace = namespace

	for _, identity := range exportNamespacesObjects {
		go func() {
			dest := squallmodels.ContentIdentifiableForIdentity(identity.Name)

			if err := manipulator.RetrieveMany(mctx, dest); err != nil {
				errorsChannel <- err
			}

			identifiablesChannel <- dest.List()
		}()

		select {
		case err := <-errorsChannel:
			return nil, err
		case ids := <-identifiablesChannel:
			identifiables = append(identifiables, ids...)
		}
	}

	return identifiables, nil
}

func TreeContentOfNamespace(namespace string, identifiables elemental.IdentifiablesList) (map[string]interface{}, error) {
	ns := &squallmodels.Namespace{}
	ns.Name = namespace

	namespaceContentRegistry := map[string][]map[string]map[string]interface{}{}
	identifiables = append(identifiables, ns)
	root := map[string]interface{}{}

	for _, identifiable := range identifiables {
		b, err := json.Marshal(identifiable)

		if err != nil {
			return nil, err
		}

		object := make(map[string]interface{})

		if err := json.Unmarshal(b, &object); err != nil {
			return nil, err
		}

		objectNamespace := reflect.ValueOf(identifiable).Elem().FieldByName("Namespace").String()

		if objectNamespace == namespace {
			objectNamespace = namespace[strings.LastIndex(namespace, "/")+1:]
		} else {
			objectNamespace = namespace[strings.LastIndex(namespace, "/")+1:] + strings.Replace(objectNamespace, namespace, "", 1)
		}

		attributeSpecifications := identifiable.(elemental.AttributeSpecifiable).AttributeSpecifications()
		FilterResourceField(attributeSpecifications, object)
		computeNamespaceAttributes(namespace, identifiable.Identity().Name, object)

		if ns == identifiable {
			root = object
		} else {
			namespaceContentRegistry[objectNamespace] = append(namespaceContentRegistry[strings.Replace(objectNamespace, namespace, "", 1)], map[string]map[string]interface{}{identifiable.Identity().Name: object})
		}
	}

	fillTreeNamespaceContent("", root, namespaceContentRegistry)
	return map[string]interface{}{"namespace": root}, nil
}

func fillTreeNamespaceContent(namespace string, currentNamespace map[string]interface{}, namespaceContentRegistry map[string][]map[string]map[string]interface{}) {

	currentNamespace["resources"] = map[string][]map[string]interface{}{}
	fullNamespaceName := namespace + currentNamespace["name"].(string)

	for _, objects := range namespaceContentRegistry[fullNamespaceName] {

		for identity, object := range objects {

			if identity == squallmodels.NamespaceIdentity.Name {
				fillTreeNamespaceContent(fullNamespaceName+"/", object, namespaceContentRegistry)
			}

			currentNamespace["resources"].(map[string][]map[string]interface{})[identity] = append(currentNamespace["resources"].(map[string][]map[string]interface{})[identity], object)
		}
	}
}

func computeNamespace(namespace string, objectNamespace string) string {
	if objectNamespace == namespace {
		return namespace[strings.LastIndex(namespace, "/"):]
	} else {
		return namespace[strings.LastIndex(namespace, "/"):] + strings.Replace(objectNamespace, namespace, "", 1)
	}
}

func computeNamespaceAttributes(namespace string, identityName string, object map[string]interface{}) {
	if identityName == squallmodels.NamespaceIdentity.Name {
		object["name"] = object["name"].(string)[strings.LastIndex(object["name"].(string), "/")+1:]
	}

	if identityName == squallmodels.APIAuthorizationPolicyIdentity.Name {
		object["authorizedNamespace"] = computeNamespace(namespace, object["authorizedNamespace"].(string))
	}

	if identityName == squallmodels.NamespaceMappingPolicyIdentity.Name {
		object["mappedNamespace"] = computeNamespace(namespace, object["mappedNamespace"].(string))
	}

	keys := []string{"subject", "object"}

	for _, key := range keys {
		if values, ok := object[key]; ok {
			for _, vs := range values.([]interface{}) {
				for i, v := range vs.([]interface{}) {
					s := strings.SplitN(v.(string), "=", 2)

					if s[0] == "$namespace" {
						newNamespace := computeNamespace(namespace, s[1])
						vs.([]interface{})[i] = s[0] + "=" + newNamespace
					}
				}
			}
		}
	}

}
