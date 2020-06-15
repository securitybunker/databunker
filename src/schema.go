package main

import (
  "context"
  "errors"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "os"
  "path/filepath"
  "strings"
  "strconv"

  "github.com/paranoidguy/jsonschema"
  jsonpatch "github.com/evanphx/json-patch"
  jptr "github.com/qri-io/jsonpointer"
)

var userSchema *jsonschema.Schema

// our custom validator
type IsLocked bool
type IsAdmin bool

func loadUserSchema(cfg Config, confFile *string) error {
  fileSchema := cfg.Generic.UserRecordSchema
  parentDir := ""
  if confFile != nil && len(*confFile) > 0 {
    parentDir = filepath.Base(*confFile)
    if parentDir != "." {
      parentDir = ""
    }
  }
  if len(fileSchema) == 0 {
    return nil
  }
  if strings.HasPrefix(fileSchema, "./") {
    _, err := os.Stat(cfg.Generic.UserRecordSchema)
    if os.IsNotExist(err) && confFile != nil {
      fileSchema = parentDir + fileSchema[2:]
	}
  } else {
    fileSchema = parentDir + fileSchema
  }
  _, err := os.Stat(fileSchema)
  if os.IsNotExist(err) {
    return err
  }
  schemaData, err := ioutil.ReadFile(fileSchema)
  if err != nil {
    return err
  }
  rs := &jsonschema.Schema{}
  jsonschema.LoadDraft2019_09()
  jsonschema.RegisterKeyword("locked", newIsLocked)
  jsonschema.RegisterKeyword("admin", newIsAdmin)
  err = rs.UnmarshalJSON(schemaData)
  if err != nil {
    return err
  }
  userSchema = rs
  return nil
}

func UserSchemaEnabled() bool {
  if userSchema == nil {
    return false
  }
  return true
}

func validateUserRecord(record []byte) error {
  if userSchema == nil {
    return nil
  }
  var doc interface{}
  if err := json.Unmarshal(record, &doc); err != nil {
    return err
  }
  result := userSchema.Validate(nil, doc)
  if len(*result.Errs) > 0 {
    return (*result.Errs)[0]
  }
  return nil
}

func validateUserRecordChange(oldRecord []byte, newRecord []byte, authResult string) (bool, error) {
  if userSchema == nil {
    return false, nil
  }
  var oldDoc interface{}
  var newDoc interface{}
  if err := json.Unmarshal(oldRecord, &oldDoc); err != nil {
    return false, err
  }
  if err := json.Unmarshal(newRecord, &newDoc); err != nil {
    return false, err
  }
  result := userSchema.Validate(nil, newDoc)
  //if len(*result.Errs) > 0 {
  //  return (*result.Errs)[0]
  //}
  result2 := userSchema.Validate(nil, oldDoc)
  if len(*result2.Errs) > 0 {
    return false, (*result.Errs)[0]
  }
  if result.ExtendedResults == nil {
    return false, nil
  }
  adminRecordChanged := false
  for _, r := range *result.ExtendedResults {
    fmt.Printf("path: %s key: %s data: %v\n", r.PropertyPath, r.Key, r.Value)
    if r.Key == "locked" || (r.Key == "admin" && authResult == "login" && adminRecordChanged == false) {
      pointer, _ := jptr.Parse(r.PropertyPath)
      data1, _ := pointer.Eval(oldDoc)
      data1Binary, _ := json.Marshal(data1)
      data2, _ := pointer.Eval(newDoc)
      data2Binary, _ := json.Marshal(data2)
      if !jsonpatch.Equal(data1Binary, data2Binary) {
        if r.Key == "locked" {
  	      fmt.Printf("Locked value changed. Old: %s New %s\n", data1Binary, data2Binary)
          return false, errors.New("User schema check error. Locked value changed: "+r.PropertyPath)
        } else {
		  fmt.Printf("Admin value changed. Approval required. Old: %s New %s\n", data1Binary, data2Binary)
          adminRecordChanged = true
        }
      }
    }
  }
  return adminRecordChanged, nil
}

func cleanupRecord(record []byte) []byte {
  empty := []byte("{}")
  if userSchema == nil {
    return empty
  }
  var doc interface{}
  if err := json.Unmarshal(record, &doc); err != nil {
    return empty
  }
  result := userSchema.Validate(nil, doc)
  if result.ExtendedResults == nil {
    return empty
  }
  
  doc1 := make(map[string]interface{})
  doc2 := make([]interface{},1)
  nested := func(path string, data interface{}) {
    currentStr := &doc1
    currentNum := &doc2
    keys := strings.Split(path, "/")
    fmt.Printf("path: %s\n", path)
    for i,k := range keys {
      if len(k) == 0 {
        continue
      }
      if kNum, err := strconv.Atoi(k); err == nil {
        if (i+1) == len(keys) {
          (*currentNum)[kNum] = data
        } else if kNxt, err := strconv.Atoi(keys[i+1]); err == nil {
          if (*currentNum)[kNum] == nil {
            v := make([]interface{}, kNxt+1)
            (*currentNum)[kNum] = v
            currentNum = &v
          } else {
            v := (*currentNum)[kNum].([]interface{})
            for (len(v) < kNxt+1) {
              v = append(v, nil)
            }
            (*currentNum)[kNum] = v
            currentNum = &v
          }
        } else {
          if (*currentNum)[kNum] == nil {
            v := make(map[string]interface{})
            (*currentNum)[kNum] = v
            currentStr = &v
          } else {
            v := (*currentNum)[kNum].(map[string]interface{})
            currentStr = &v
          }
        }
      } else {
        if (i+1) == len(keys) {
          (*currentStr)[k] = data
        } else if kNxt, err := strconv.Atoi(keys[i+1]); err == nil {
          if _, ok := (*currentStr)[k]; !ok {
            v := make([]interface{}, kNxt+1)
            (*currentStr)[k] = v
            currentNum = &v
          } else {
            v := (*currentStr)[k].([]interface{})
            for (len(v) < kNxt+1) {
              v = append(v, nil)
            }
            (*currentStr)[k] = v
            currentNum = &v
          }
        } else {
          if _, ok := (*currentStr)[k]; !ok {
            v := make(map[string]interface{})
            (*currentStr)[k] = v
            currentStr = &v
          } else {
            v := (*currentNum)[kNum].(map[string]interface{})
            currentStr = &v
          }
        }
      }
    }
  }

  for _, r := range *result.ExtendedResults {
    fmt.Printf("path: %s key: %s data: %v\n", r.PropertyPath, r.Key, r.Value)
    if r.Key == "leftover" {
      //pointer, _ := jptr.Parse(r.PropertyPath)
      //data1, _ := pointer.Eval(oldDoc)
      nested(r.PropertyPath, r.Value)
    }
  }
  fmt.Printf("final doc1 %v\n", doc1)
  dataBinary, err := json.Marshal(doc1)
  fmt.Println(err)
  fmt.Printf("data bin %s\n", dataBinary)
  return dataBinary
}

// Locked keyword - meaningin that value should never be changed after record created
func newIsLocked() jsonschema.Keyword {
  return new(IsLocked)
}

// Validate implements jsonschema.Keyword
func (f *IsLocked) Validate(propPath string, data interface{}, errs *[]jsonschema.KeyError) {
  fmt.Printf("Validate: %s -> %v\n", propPath, data)
}

// Register implements jsonschema.Keyword
func (f *IsLocked) Register(uri string, registry *jsonschema.SchemaRegistry) {
}

// Resolve implements jsonschema.Keyword
func (f *IsLocked) Resolve(pointer jptr.Pointer, uri string) *jsonschema.Schema {
  fmt.Printf("Resolve %s\n", uri)
  return nil
}

func (f *IsLocked) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
  fmt.Printf("ValidateKeyword locked %s => %v\n", currentState.InstanceLocation.String(), data)
  currentState.AddExtendedResult("locked", data)
}

// Admin keyword. Any change in this record requires admin approval.
func newIsAdmin() jsonschema.Keyword {
  return new(IsAdmin)
}

// Validate implements jsonschema.Keyword
func (f *IsAdmin) Validate(propPath string, data interface{}, errs *[]jsonschema.KeyError) {
  fmt.Printf("Validate: %s -> %v\n", propPath, data)
}

// Register implements jsonschema.Keyword
func (f *IsAdmin) Register(uri string, registry *jsonschema.SchemaRegistry) {
}

// Resolve implements jsonschema.Keyword
func (f *IsAdmin) Resolve(pointer jptr.Pointer, uri string) *jsonschema.Schema {
  fmt.Printf("Resolve %s\n", uri)
  return nil
}

func (f *IsAdmin) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
  fmt.Printf("ValidateKeyword admin %s => %v\n", currentState.InstanceLocation.String(), data)
  currentState.AddExtendedResult("admin", data)
}
