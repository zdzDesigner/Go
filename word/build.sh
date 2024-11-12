#!/bin/sh



# GOARCH=wasm GOOS=js go build -o main.wasm main.go


 go test -v ./lib -count=1 -test.run Test_Docx2md



# f.Name: docProps/app.xml
# f.Name: _rels/.rels
# f.Name: docProps/core.xml
# f.Name: word/document.xml
# f.Name: word/endnotes.xml
# f.Name: word/_rels/document.xml.rels
# f.Name: customXml/item3.xml
# f.Name: customXml/itemProps31.xml
# f.Name: customXml/_rels/item3.xml.rels
# f.Name: word/footnotes.xml
# f.Name: customXml/item22.xml
# f.Name: customXml/itemProps22.xml
# f.Name: customXml/_rels/item22.xml.rels
# f.Name: customXml/item13.xml
# f.Name: customXml/itemProps13.xml
# f.Name: customXml/_rels/item13.xml.rels
# f.Name: word/webSettings.xml
# f.Name: word/theme/theme11.xml
# f.Name: word/settings.xml
# f.Name: word/glossary/document.xml
# f.Name: word/glossary/webSettings2.xml
# f.Name: word/glossary/_rels/document.xml.rels
# f.Name: word/glossary/settings2.xml
# f.Name: word/glossary/styles.xml
# f.Name: word/glossary/fontTable.xml
# f.Name: word/styles2.xml
# f.Name: word/fontTable2.xml
# f.Name: docProps/custom.xml
# f.Name: [Content_Types].xml
# word/document*.xml docProps/app.xml
# word/document*.xml _rels/.rels
# word/document*.xml docProps/core.xml
# word/document*.xml word/document.xml

