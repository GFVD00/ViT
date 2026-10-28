package doc

import (
	"fmt"
	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

func Create(text, filePath string) {
	// 1. Open existing DOCX document[cite: 6]
	doc, err := docx.OpenDocument(filePath)
	if err != nil {
		fmt.Println("Error opening docx file:", err)
		return
	}

	// 2. Create a new paragraph[cite: 6]
	para, err := doc.AddParagraph()
	if err != nil {
		fmt.Println("Error adding paragraph:", err)
		return
	}

	// 3. Create a text run[cite: 6]
	run, err := para.AddRun()
	if err != nil {
		fmt.Println("Error adding text run:", err)
		return
	}

	// 4. Set text[cite: 6]
	run.SetText(text)

	// 5. Set font family and size[cite: 6]
	// Note: In MS Word XML (OOXML), font size is measured in half-points[cite: 6].
	// Therefore, for 14 pt you need to pass 14 * 2 = 28[cite: 6].
	run.SetFont(domain.Font{
		Name: "Times New Roman",
	})
	run.SetSize(28)

	// 6. Save changes to file[cite: 6]
	err = doc.SaveAs(filePath)
	if err != nil {
		fmt.Println("Error saving docx file:", err)
	}
}
