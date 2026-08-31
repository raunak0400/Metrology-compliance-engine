# Placeholder image for the separate Python ML service.
# The real service (OpenCV / PaddleOCR / custom detection + NER models) lives on
# the `ml-service` branch and is owned by the ML team. This stub only exposes the
# HTTP contract the Node backend orchestrates against, so the stack runs locally.
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY app.py ./
EXPOSE 8000
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
