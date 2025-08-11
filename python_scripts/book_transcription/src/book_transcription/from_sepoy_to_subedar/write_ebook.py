from ebooklib import epub
import uuid
from pathlib import Path

def _add_authors(book: epub.EpubBook):
    book.add_author(
        "Sita Ram Pandey",
        file_as="Sita Ram Pandey",
        role="author",
        uid="author",
    )
    book.add_author(
        "Lieutenant-Colonel Norgate",
        file_as="Lieutenant-Colonel Norgate",
        role="translator",
        uid="translator",
    )
    book.add_author(
        "James Lunt",
        file_as="James Lunt",
        role="editor",
        uid="editor",
    )
    book.add_author(
        "Frank Wilson",
        file_as="Frank Wilson",
        role="illustrator",
        uid="illustrator",
    )

if __name__ == "__main__":
    current_dir = Path(__file__).parent
    raw_content_dir = current_dir / "raw_content"
    content_paths = {}
    contents = {}
    book_parts = ["title_page",
                  "dedication",
                  "translator_description",
                  "preface_by_translator",
                  "editorial_note",
                  "acknowledgements",
                  "introduction",
                  "foreward_by_sita_ram",
                  "beginning",
                  "joining_the_regiment",
                  "the_gurkha_war",
                  "the_pindari_war",
                  "return_to_the_village",
                  "the_lovely_thakurin",
                  "the_bulwark_of_hindustan",
                  "the_march_into_afghanistan",
                  "ghazni_and_kabul",
                  "the_retreat_from_kabul",
                  "escape_from_slavery",
                  "the_first_sikh_war",
                  "the_second_sikh_war",
                  "the_wind_of_madness",
                  "the_pensioner"
                ]
    book_part_titles = {
        "title_page": "Title Page",
        "dedication": "Dedication",
        "translator_description": "Translator's Description",
        "preface_by_translator": "Preface by the Translator",
        "editorial_note": "Editorial Note",
        "acknowledgements": "Acknowledgements",
        "introduction": "Introduction",
        "foreward_by_sita_ram": "Foreword by Sita Ram",
        "beginning": "The Beginning",
        "joining_the_regiment": "Joining the Regiment",
        "the_gurkha_war": "The Gurkha War: 1814 - 1816",
        "the_pindari_war": "The Pindari War",
        "return_to_the_village": "Return to the Village",
        "the_lovely_thakurin": "The Lovely Thakurin",
        "the_bulwark_of_hindustan": "The Bulwark of Hindustan",
        "the_march_into_afghanistan": "The March into Afghanistan: 1838-1839",
        "ghazni_and_kabul": "Ghazni and Kabul",
        "the_retreat_from_kabul": "The Retreat from Kabul: January 1842",
        "escape_from_slavery": "Escape from Slavery",
        "the_first_sikh_war": "The First Sikh War: 1845-1846",
        "the_second_sikh_war": "The Second Sikh War: 1848-1849",
        "the_wind_of_madness": "The Wind of Madness",
        "the_pensioner": "The Pensioner"
    }
    for part in book_parts:
        content_paths[part] = raw_content_dir / f"{part}.xhtml"

    for content_name, content_path in content_paths.items():
        assert content_path.exists(), f"Content path {content_path} does not exist"
        with open(content_path, "r") as f:
            contents[content_name] = str(f.read())

    book = epub.EpubBook()

    # set metadata
    book.set_identifier(str(uuid.uuid4()))
    book.set_title("From Sepoy to Subedar")
    book.set_language("en")
    _add_authors(book)
    content_items = []
    for content_name, content in contents.items():
        content_item = epub.EpubHtml(title=book_part_titles[content_name], file_name=f"{content_name}.xhtml", lang="en")
        content_item.content = content
        book.add_item(content_item)
        content_items.append(content_item)

    # Add a TOC
    book.toc = (
        epub.Section("From Sepoy to Subedar"),
        *content_items,
    )

    style = "BODY {color: white;}"
    nav_css = epub.EpubItem(
        uid="style_nav",
        file_name="style/nav.css",
        media_type="text/css",
        content=style,
    )
    # add default NCX and Nav file
    book.add_item(epub.EpubNcx())
    book.add_item(epub.EpubNav())

    # add CSS file
    book.add_item(nav_css)

    # basic spine
    book.spine = ["nav", *content_items]

    # write to the file
    filepath = Path("from_sepoy_to_subedar.epub")
    if filepath.exists():
        filepath.unlink()
    epub.write_epub("from_sepoy_to_subedar.epub", book, {})