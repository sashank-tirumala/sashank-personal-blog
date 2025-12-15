from ebooklib import epub
import uuid
from pathlib import Path

def _add_authors(book: epub.EpubBook):
    book.add_author(
        "Vishnu Sitaram Sukthankar",
        file_as="Vishnu Sitaram Sukhthankar",
        role="author",
        uid="author",
    )

if __name__ == "__main__":
    current_dir = Path(__file__).parent
    raw_content_dir = current_dir / "raw_content"
    content_paths = {}
    contents = {}
    book_parts =[
        "introductory_note",
        "lecture_1",
        "lecture_2",
        "lecture_3",
        "lecture_4"
                ]
    book_part_titles = {
        "introductory_note": "Introductory Note",
        "lecture_1": "Lecture I: The Mahabharatha and it's Critics",
        "lecture_2": "Lecture II: The Story on the Mundane Plane",
        "lecture_3": "Lecture III: The Story on the Ethical Plane",
        "lecture_4": "Lecture IV: The Story on the Metaphysical Plane"
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
    book.set_title("On the meaning of the Mahabharatha")
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
        epub.Section("On the meaning of the Mahabharatha"),
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
    filepath = Path("on_the_meaning_of_mahabharatha.epub")
    if filepath.exists():
        filepath.unlink()
    epub.write_epub("on_the_meaning_of_mahabharatha.epub", book, {})