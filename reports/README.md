# Reports

- **reportHI1.pdf** – Hand-in 1 (Peer-to-Peer Flooding and Simple Ledger).
- **reportHI3.pdf** – Hand-in 3 (Validation and measurements).

Sources for Hand-in 3: `report.md`, `report-header.tex`. To regenerate the PDF from the repo root:

```bash
pandoc reports/report.md -o reports/reportHI3.pdf --pdf-engine=pdflatex -H reports/report-header.tex -V colorlinks=true
```
