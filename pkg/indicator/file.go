package indicator

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "log"
)

func ReadFile(indicatorsFile string, overrideMetadata ...map[string]string) (Document, error) {
  fileBytes, err := ioutil.ReadFile(indicatorsFile)
  if err != nil {
    return Document{}, err
  }

  indicatorDocument, err := ReadIndicatorDocument(fileBytes)
  if err != nil {
    return Document{}, err
  }

  indicatorDocument, err = fillInMetadata(indicatorDocument.Metadata, overrideMetadata, fileBytes)
  if err != nil {
    return Document{}, err
  }

  validationErrors := Validate(indicatorDocument)
  if len(validationErrors) > 0 {

    for _, e := range validationErrors {
      log.Printf("- %s \n", e.Error())
    }

    return Document{}, fmt.Errorf("validation for indicator file failed - [%+v]", validationErrors)
  }

  return indicatorDocument, nil
}

func fillInMetadata(documentMetadata map[string]string, overrideMetadata []map[string]string, documentBytes []byte) (Document, error) {

  for _, overrides := range overrideMetadata {
    for k, v := range overrides {
      documentMetadata[k] = v
    }
  }

  for k, v := range documentMetadata {
    documentBytes = bytes.Replace(documentBytes, []byte("$"+k), []byte(v), -1)
  }

  indicatorDocument, err := ReadIndicatorDocument(documentBytes)
  if err != nil {
    return Document{}, err
  }

  return indicatorDocument, nil
}
