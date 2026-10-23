package contractcheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// contract holds this repository's descriptors and its access option.
type contract struct {
	files  []*descriptorpb.FileDescriptorProto
	access protoreflect.ExtensionType
}

// loadContract builds descriptors and resolves the custom method option.
func loadContract(t *testing.T) contract {
	t.Helper()
	t.Chdir("../../proto")
	image := filepath.Join(t.TempDir(), "image.binpb")
	if output, err := buf(t, 2*time.Minute, "build", "-o", image); err != nil {
		t.Fatalf("build the contract: %v\n%s", err, output)
	}
	raw, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	var set descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(raw, &set); err != nil {
		t.Fatal(err)
	}
	files, err := protodesc.NewFiles(&set)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := files.FindDescriptorByName("common.v1.access")
	if err != nil {
		t.Fatal(err)
	}
	extension := dynamicpb.NewExtensionType(descriptor.(protoreflect.ExtensionDescriptor))
	resolver := new(protoregistry.Types)
	if err := resolver.RegisterExtension(extension); err != nil {
		t.Fatal(err)
	}
	if err := (proto.UnmarshalOptions{Resolver: resolver}).Unmarshal(raw, &set); err != nil {
		t.Fatal(err)
	}
	built := contract{access: extension}
	for _, file := range set.GetFile() {
		if strings.HasPrefix(file.GetName(), "google/") {
			continue
		}
		built.files = append(built.files, file)
	}
	if len(built.files) == 0 {
		t.Fatal("the image holds no files of this repository")
	}
	return built
}

// messages visits every message of a file, nested ones included, with its qualified name.
func messages(file *descriptorpb.FileDescriptorProto, visit func(name string, message *descriptorpb.DescriptorProto)) {
	var walk func(prefix string, list []*descriptorpb.DescriptorProto)
	walk = func(prefix string, list []*descriptorpb.DescriptorProto) {
		for _, message := range list {
			name := prefix + "." + message.GetName()
			if !message.GetOptions().GetMapEntry() {
				visit(name, message)
			}
			walk(name, message.GetNestedType())
		}
	}
	walk(file.GetPackage(), file.GetMessageType())
}

func TestEveryRPCStatesItsAccess(t *testing.T) {
	built := loadContract(t)
	for _, file := range built.files {
		for _, service := range file.GetService() {
			for _, method := range service.GetMethod() {
				options := method.GetOptions().ProtoReflect()
				field := built.access.TypeDescriptor()
				if !options.Has(field) || options.Get(field).Enum() == 0 {
					t.Errorf("%s.%s.%s: missing explicit access option", file.GetPackage(), service.GetName(), method.GetName())
				}
			}
		}
	}
}

func TestPartialUpdatesUseFieldPresence(t *testing.T) {
	built := loadContract(t)
	for _, file := range built.files {
		messages(file, func(name string, message *descriptorpb.DescriptorProto) {
			for _, field := range message.GetField() {
				if field.GetTypeName() == ".google.protobuf.FieldMask" {
					t.Errorf("%s.%s: use optional fields for partial updates, not a field mask", name, field.GetName())
				}
			}
		})
	}
}

func TestListsPageTheSameWay(t *testing.T) {
	built := loadContract(t)
	fields := map[string]map[string]*descriptorpb.FieldDescriptorProto{}
	for _, file := range built.files {
		messages(file, func(name string, message *descriptorpb.DescriptorProto) {
			fields[name] = map[string]*descriptorpb.FieldDescriptorProto{}
			for _, field := range message.GetField() {
				fields[name][field.GetName()] = field
			}
		})
	}
	for _, file := range built.files {
		for _, service := range file.GetService() {
			for _, method := range service.GetMethod() {
				request := fields[strings.TrimPrefix(method.GetInputType(), ".")]
				response := fields[strings.TrimPrefix(method.GetOutputType(), ".")]
				where := file.GetPackage() + "." + service.GetName() + "." + method.GetName()
				if _, pages := request["page_token"]; pages {
					if request["page_size"].GetType() != descriptorpb.FieldDescriptorProto_TYPE_INT32 {
						t.Errorf("%s: a paged request needs an int32 page_size", where)
					}
					if response["next_page_token"].GetType() != descriptorpb.FieldDescriptorProto_TYPE_STRING {
						t.Errorf("%s: a paged response needs a string next_page_token", where)
					}
				}
				if _, pages := response["next_page_token"]; pages {
					if _, takes := request["page_token"]; !takes {
						t.Errorf("%s: a response with next_page_token needs a request with page_token", where)
					}
				}
			}
		}
	}
}

// Open sets are strings on purpose; each is listed with the reason.
var openVocabularies = map[string]string{
	"audit.v1.AuditEntry.target_type":  "the kinds of thing the journal records grow with the product",
	"billing.v1.Payment.cancel_reason": "the provider's own reason",
	"history.v1.Record.content_type":   "a media type",
}

var vocabularyName = regexp.MustCompile(`(^|_)(status|kind|state|scope|level|side|source|mode|visibility|theme|period|reason|type)$`)

func TestClosedVocabulariesAreEnums(t *testing.T) {
	built := loadContract(t)
	for _, file := range built.files {
		messages(file, func(name string, message *descriptorpb.DescriptorProto) {
			for _, field := range message.GetField() {
				if field.GetType() != descriptorpb.FieldDescriptorProto_TYPE_STRING || !vocabularyName.MatchString(field.GetName()) {
					continue
				}
				qualified := name + "." + field.GetName()
				if _, open := openVocabularies[qualified]; !open {
					t.Errorf("%s: a string named like a vocabulary; make it an enum, or list it as an open set here", qualified)
				}
			}
		})
	}
}

// One spelling and one word for each thing, in names (comments follow the same terms by review).
var wrongNames = map[string]string{
	"colour":    "color",
	"cancelled": "canceled",
	"space":     "workspace",
}

func TestNamesUseOneSpellingAndOneTerm(t *testing.T) {
	built := loadContract(t)
	check := func(kind, qualified, name string) {
		for _, word := range words(name) {
			if right, wrong := wrongNames[word]; wrong {
				t.Errorf("%s %s: use %q, not %q", kind, qualified, right, word)
			}
		}
	}
	for _, file := range built.files {
		messages(file, func(name string, message *descriptorpb.DescriptorProto) {
			check("message", name, message.GetName())
			for _, field := range message.GetField() {
				check("field", name+"."+field.GetName(), field.GetName())
			}
		})
		for _, enum := range file.GetEnumType() {
			check("enum", file.GetPackage()+"."+enum.GetName(), enum.GetName())
			for _, value := range enum.GetValue() {
				check("enum value", file.GetPackage()+"."+value.GetName(), value.GetName())
			}
		}
		for _, service := range file.GetService() {
			check("service", file.GetPackage()+"."+service.GetName(), service.GetName())
			for _, method := range service.GetMethod() {
				check("method", file.GetPackage()+"."+service.GetName()+"."+method.GetName(), method.GetName())
			}
		}
	}
}

// words splits a CamelCase or SCREAMING_SNAKE identifier into lower-case words.
func words(identifier string) []string {
	var out []string
	var current []rune
	flush := func() {
		if len(current) > 0 {
			out = append(out, string(current))
			current = nil
		}
	}
	runes := []rune(identifier)
	for i, r := range runes {
		switch {
		case r == '_':
			flush()
		case unicode.IsUpper(r) && i > 0 && unicode.IsLower(runes[i-1]):
			flush()
			current = append(current, unicode.ToLower(r))
		default:
			current = append(current, unicode.ToLower(r))
		}
	}
	flush()
	return out
}
