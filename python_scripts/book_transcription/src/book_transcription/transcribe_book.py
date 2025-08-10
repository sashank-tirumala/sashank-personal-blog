from mistralai import Mistral
import os

api_key = os.environ["MISTRAL_API_KEY"]

client = Mistral(api_key=api_key)

uploaded_pdf = client.files.upload(
    file={
        "file_name": "uploaded_file.pdf",
        "content": open("/Users/sashanktirumala/Desktop/Introduction.pdf", "rb"),
    },
    purpose="ocr"
)
signed_url = client.files.get_signed_url(file_id=uploaded_pdf.id)

ocr_response = client.ocr.process(
    model="mistral-ocr-latest",
    document={
        "type": "document_url",
        "document_url": signed_url.url,
    },
    include_image_base64=True
)
with open("ocr_output.json", "w") as f:
    f.write(ocr_response.json())
