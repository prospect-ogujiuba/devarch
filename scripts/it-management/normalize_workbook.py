#!/usr/bin/env python3
"""Normalize the IT management XLSM workbook without modifying it."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
from pathlib import Path
from zipfile import ZipFile
import xml.etree.ElementTree as ET

MAIN = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
REL = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
PKG_REL = "http://schemas.openxmlformats.org/package/2006/relationships"
DATE_FORMAT_IDS = {14, 15, 16, 17, 18, 19, 20, 21, 22, 27, 30, 36, 45, 46, 47, 50, 57}
DATA_SHEETS = ("Tickets", "Archive", "Employees", "Printers", "Provisions", "DataTables")


def column_number(reference: str) -> int:
    letters = re.match(r"[A-Z]+", reference).group(0)
    result = 0
    for char in letters:
        result = result * 26 + ord(char) - 64
    return result


def excel_date(value: float) -> str:
    value = float(value)
    base = dt.datetime(1899, 12, 30)
    parsed = base + dt.timedelta(days=value)
    if parsed.time() == dt.time():
        return parsed.date().isoformat()
    return parsed.isoformat(timespec="seconds")


def normalize_scalar(value):
    if isinstance(value, str):
        return value.strip()
    return value


def read_workbook(path: Path) -> dict[str, list[dict]]:
    with ZipFile(path) as archive:
        def xml(name: str):
            return ET.fromstring(archive.read(name))

        shared_strings = []
        if "xl/sharedStrings.xml" in archive.namelist():
            for item in xml("xl/sharedStrings.xml"):
                shared_strings.append("".join(node.text or "" for node in item.iter(f"{{{MAIN}}}t")))

        date_styles = set()
        if "xl/styles.xml" in archive.namelist():
            styles = xml("xl/styles.xml")
            custom_dates = set()
            num_fmts = styles.find(f"{{{MAIN}}}numFmts")
            if num_fmts is not None:
                for fmt in num_fmts:
                    code = fmt.attrib.get("formatCode", "").lower()
                    if re.search(r"(^|[^\\])[ymdhis]", code):
                        custom_dates.add(int(fmt.attrib["numFmtId"]))
            cell_xfs = styles.find(f"{{{MAIN}}}cellXfs")
            if cell_xfs is not None:
                for index, xf in enumerate(cell_xfs):
                    fmt_id = int(xf.attrib.get("numFmtId", 0))
                    if fmt_id in DATE_FORMAT_IDS or fmt_id in custom_dates:
                        date_styles.add(index)

        relationships = {
            node.attrib["Id"]: node.attrib["Target"]
            for node in xml("xl/_rels/workbook.xml.rels").findall(f"{{{PKG_REL}}}Relationship")
        }
        workbook = xml("xl/workbook.xml")
        sheets = {}

        for sheet in workbook.find(f"{{{MAIN}}}sheets"):
            title = sheet.attrib["name"]
            if title not in DATA_SHEETS:
                continue
            target = relationships[sheet.attrib[f"{{{REL}}}id"]].lstrip("/")
            if not target.startswith("xl/"):
                target = f"xl/{target}"
            root = xml(target)
            raw_rows = []
            for row in root.findall(f".//{{{MAIN}}}sheetData/{{{MAIN}}}row"):
                values = {}
                for cell in row.findall(f"{{{MAIN}}}c"):
                    index = column_number(cell.attrib["r"])
                    cell_type = cell.attrib.get("t")
                    value_node = cell.find(f"{{{MAIN}}}v")
                    if cell_type == "s" and value_node is not None:
                        value = shared_strings[int(value_node.text)]
                    elif cell_type == "inlineStr":
                        value = "".join(node.text or "" for node in cell.iter(f"{{{MAIN}}}t"))
                    elif cell_type == "b" and value_node is not None:
                        value = value_node.text == "1"
                    elif value_node is not None:
                        raw = value_node.text or ""
                        try:
                            number = float(raw)
                            value = int(number) if number.is_integer() else number
                            style = int(cell.attrib.get("s", 0))
                            if style in date_styles:
                                value = excel_date(number)
                        except ValueError:
                            value = raw
                    else:
                        value = ""
                    values[index] = normalize_scalar(value)
                if any(value not in (None, "") for value in values.values()):
                    raw_rows.append(values)

            if not raw_rows:
                sheets[title] = []
                continue
            header_row = raw_rows[0]
            headers = {index: str(value).strip() for index, value in header_row.items() if value not in (None, "")}
            records = []
            for row in raw_rows[1:]:
                record = {header: row.get(index, "") for index, header in headers.items()}
                if any(value not in (None, "") for value in record.values()):
                    records.append(record)
            sheets[title] = records

        return sheets


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("workbook", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    data = read_workbook(args.workbook)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8")
    args.output.chmod(0o600)
    print(json.dumps({name: len(rows) for name, rows in data.items()}, sort_keys=True))


if __name__ == "__main__":
    main()
