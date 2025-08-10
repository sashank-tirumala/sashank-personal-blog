from ebooklib import epub
import uuid
from pathlib import Path

book = epub.EpubBook()

# set metadata
book.set_identifier(str(uuid.uuid4()))
book.set_title("From Sepoy to Subedar")
book.set_language("en")

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
    uid="coauthor",
)
book.add_author(
    "James Lunt",
    file_as="James Lunt",
    role="editor",
    uid="coauthor",
)
book.add_author(
    "Frank Wilson",
    file_as="Frank Wilson",
    role="illustrator",
    uid="coauthor",
)

# Create dedication page
dedication = epub.EpubHtml(title="Dedication", file_name="dedication.xhtml", lang="en")
dedication.content = (
    """
    <p class="lead">This book is dedicated to the</p>
    <p class="name">jawan</p>
    <p class="sub">past and present</p>
    <p class="closing">with admiration and affection</p>
    """
)
book.add_item(dedication)

# Create title page
title_page = epub.EpubHtml(title="Title Page", file_name="title.xhtml", lang="en")
title_page.content = (
    """
    <div class="title-page">
        <h1 class="book-title">From Sepoy to Subedar</h1>
        <div class="author-info">
            <p class="author">by Sita Ram Pandey</p>
            <p class="translator">Translated by Lieutenant-Colonel Norgate</p>
            <p class="editor">Edited by James Lunt</p>
            <p class="illustrator">Illustrated by Frank Wilson</p>
        </div>
    </div>
    """
)
book.add_item(title_page)

# create chapter
c1 = epub.EpubHtml(title="Intro", file_name="chap_01.xhtml", lang="hr")
c1.content = (
    "<h1>Intro heading</h1>"
    "<p>Zaba je skocila u baru.</p>"
    '<p><img alt="[ebook logo]" src="static/ebooklib.gif"/><br/></p>'
)

# create image from the local image
# image_content = open("ebooklib.gif", "rb").read()
# img = epub.EpubImage(
#     uid="image_1",
#     file_name="static/ebooklib.gif",
#     media_type="image/gif",
#     content=image_content,
# )

# add chapter
book.add_item(c1)
# add image
# book.add_item(img)

# define Table Of Contents
# book.toc = (
#     epub.Link("chap_01.xhtml", "Introduction", "intro"),
#     (epub.Section("Simple book"), (c1,)),
# )

# add default NCX and Nav file
book.add_item(epub.EpubNcx())
book.add_item(epub.EpubNav())

# define CSS style
style = "BODY {color: white;}"
nav_css = epub.EpubItem(
    uid="style_nav",
    file_name="style/nav.css",
    media_type="text/css",
    content=style,
)

# add CSS file
book.add_item(nav_css)

# basic spine
book.spine = ["nav", title_page, dedication]

# write to the file
filepath = Path("test.epub")
if filepath.exists():
    filepath.unlink()
epub.write_epub("test.epub", book, {})