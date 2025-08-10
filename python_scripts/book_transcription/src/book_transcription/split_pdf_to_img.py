from pathlib import Path
from pdf2image import convert_from_path
import argparse
import shutil

def split_pdf_to_png(pdf_path: Path) -> None:
    output_dir = pdf_path.parent / f"{pdf_path.stem}_images"
    if output_dir.exists():
        print(f"Output directory {output_dir} already exists. Removing it.")
        shutil.rmtree(output_dir)
        print(f"Removed existing directory {output_dir}.")
    output_dir.mkdir()

    pages = convert_from_path(str(pdf_path))
    for i, page in enumerate(pages, start=1):
        page.save(output_dir/ f"{i}.png", "PNG")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Split PDF into PNG images.")
    parser.add_argument("--pdf_path", type=Path, help="Path to the PDF file.")
    args = parser.parse_args()
    output_dir = args.pdf_path.parent / f"{args.pdf_path.stem}_images"
    
    split_pdf_to_png(args.pdf_path)
    print(f"PDF {args.pdf_path} has been split into images in {output_dir}.")
