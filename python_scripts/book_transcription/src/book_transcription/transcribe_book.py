from openai import OpenAI
from pathlib import Path
from tqdm import tqdm

client = OpenAI()

directory = Path(
    "/Users/sashanktirumala/Desktop/from_sepoy_to_subedar/the_lovely_thakurin_images"
)
images = list(directory.glob("*.png"))
images.sort(key=lambda x: int(x.stem))  # Sort images by their numeric filename
out_text = ""
for image in tqdm(images, desc="Processing images"):
    # Upload the image to OpenAI
    file = client.files.create(file=open(image, "rb"), purpose="user_data")

    response = client.responses.create(
        model="gpt-5-mini",
        input=[
            {
                "role": "developer",
                "content": [
                    {
                        "type": "input_text",
                        "text": '# Objective\nConvert a provided image of a book page into an XHTML fragment suitable for inclusion in an EPUB, faithfully representing all visible content and structure.\n\n# Instructions\n- Review the image carefully to identify all readable content, including headings, titles, footnotes, poetic lines, lists, and emphasis (e.g., italics, bold).\n- Accurately transcribe all content, inferring appropriate XHTML tags:\n  - Use `<h1>`, `<h2>`, etc. for headings/titles.\n  - Mark emphasis with `<em>` or `<strong>` as visible.\n  - Retain distinct formatting for poetry (`<blockquote>`, `<br />`, etc.), lists (`<ul>`, `<ol>`, `<li>`), or similar features.\n- For footnotes:\n  - Inline footnote references: `<a href="#footnoteX" epub:type="noteref"><sup>X</sup></a>` (with unique `X`) at the reference point. <sup> is important, dont forget. \n  - Footnote content at end: `<aside id="footnoteX" epub:type="footnote">\u00026hellip;</aside>`, assigning matching IDs.\n- Paragraph tags:\n  - If a paragraph starts on this page: `<p>` at the start, close only if the paragraph concludes.\n  - If the text begins/ends in the middle of a paragraph, omit opening/closing `<p>` as appropriate.\n- If content is unclear or ambiguous, insert `<!-- [UNCLEAR] -->` at the precise spot.\n- Do NOT add markup for page numbers, boundaries, or continuation.\n- Do NOT include `<html>`, `<head>`, or `<body>` tags—return only the core content fragment.\n Do NOT include any data from a page that only contains a map \n Do not include the small page header that indicates the chapter name, large chpater title on a page is fine to include\n\nAfter transcription, validate that all key content types and structural elements have been appropriately tagged according to EPUB standards. If any ambiguity or unhandled case is detected, self-correct or flag the location with an inline comment as specified.\n\n# Output Format\n- Return only the relevant XHTML fragment as plain text, nothing else.\n- Footnotes must be listed at the end using proper EPUB markup as above.\n- If any text or formatting is uncertain, use `<!-- [UNCLEAR] -->` inline at that location.\n\n# Process Steps\n1. Analyze the image for structure and formatting cues (titles, headings, footnotes, lists, poetry, emphasized text).\n2. Assign correct XHTML tags for each structural or semantic element, ensuring EPUB compliance.\n3. Transcribe the page content, preserving original formatting and structure.\n4. Implement EPUB-compliant footnote references and their content blocks.\n5. Use opening or closing `<p>` tags only when a paragraph fully starts or ends on the page.\n\n# Example\n_Input (page contains):_\n  - Ending of a paragraph: “…wonderful discovery.<sup>1</sup></p>”\n  - Chapter title: “CHAPTER IV”\n  - Start of a paragraph: “This new phase was unexpected and—”\n\n_Output:_\n…wonderful discovery.<a href="#footnote1" epub:type="noteref">1</a></p>\n<h1>CHAPTER IV</h1>\n<p>This new phase was unexpected and—\n\n<aside id="footnote1" epub:type="footnote">\n    1. This is the corresponding footnote text.\n</aside>\n\n(For unclear areas, insert: `<!-- [UNCLEAR] -->` at the location.)\n\n# Reminder\nOnly return the XHTML content fragment for the book page. Do not include any explanations, wrappers, or non-content text.',
                    }
                ],
            },
            {
                "role": "user",
                "content": [
                    {
                        "type": "input_image",
                        "file_id": file.id,
                    }
                ],
            },
        ],
        text={"format": {"type": "text"}, "verbosity": "medium"},
        reasoning={"effort": "medium", "summary": None},
        tools=[],
        store=True,
    )
    out_text += response.output_text + "\n"

output_html_file = directory / "output.xhtml"
with open(output_html_file, "w") as f:
    f.write(out_text)

print(f"Output written to {output_html_file}")
