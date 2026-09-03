"""Import normalized IT management workbook data into Taiga using Django models."""

from __future__ import annotations

import datetime as dt
import hashlib
import json
import re
import secrets
import sys

from django.apps import apps
from django.db import transaction
from django.utils import timezone

if len(sys.argv) != 2:
    raise SystemExit("usage: manage.py shell < import_taiga.py -- NORMALIZED_JSON")

source_path = sys.argv[1]
with open(source_path, encoding="utf-8") as source:
    data = json.load(source)

User = apps.get_model("users", "User")
Role = apps.get_model("users", "Role")
Project = apps.get_model("projects", "Project")
ProjectTemplate = apps.get_model("projects", "ProjectTemplate")
Membership = apps.get_model("projects", "Membership")
Issue = apps.get_model("issues", "Issue")
IssueStatus = apps.get_model("projects", "IssueStatus")
IssueType = apps.get_model("projects", "IssueType")
Priority = apps.get_model("projects", "Priority")
IssueCustomAttribute = apps.get_model("custom_attributes", "IssueCustomAttribute")
IssueCustomAttributesValues = apps.get_model("custom_attributes", "IssueCustomAttributesValues")

PROJECT_SLUG = "it-management-system"
PROJECT_NAME = "IT Management System"


def text(value):
    return "" if value is None else str(value).strip()


def slug(value, limit=48):
    result = re.sub(r"[^a-z0-9]+", "-", text(value).lower()).strip("-")
    return result[:limit]


def valid_email(value):
    return bool(re.fullmatch(r"[^@\s]+@[^@\s]+\.[^@\s]+", text(value)))


def parse_date(value):
    try:
        return dt.date.fromisoformat(text(value)[:10])
    except (ValueError, TypeError):
        return None


def aware_datetime(value):
    date = parse_date(value)
    if not date:
        return None
    return timezone.make_aware(dt.datetime.combine(date, dt.time(hour=12)))


def normalized_choice(value):
    value = re.sub(r"^\s*\d+\s*-\s*", "", text(value)).strip()
    return value or "Unspecified"


summary = {
    "users_created": 0,
    "users_updated": 0,
    "memberships": 0,
    "department_roles": 0,
    "issues_created": 0,
    "issues_updated": 0,
    "custom_attributes": 0,
    "unresolved_people": 0,
}

with transaction.atomic():
    admin = User.objects.filter(is_superuser=True, is_active=True).first()
    if not admin:
        raise RuntimeError("Taiga has no active superuser")

    project = Project.objects.filter(slug=PROJECT_SLUG).first()
    if project is None:
        template = ProjectTemplate.objects.get(slug="kanban")
        project = Project.objects.create(
            name=PROJECT_NAME,
            slug=PROJECT_SLUG,
            description="Imported from IT_Management_System.xlsm",
            owner=admin,
            creation_template=template,
            is_private=True,
            is_backlog_activated=False,
            is_kanban_activated=True,
            is_wiki_activated=True,
            is_issues_activated=True,
        )
    else:
        project.name = PROJECT_NAME
        project.description = "Imported from IT_Management_System.xlsm"
        project.is_private = True
        project.is_issues_activated = True
        project.save()

    owner_membership = Membership.objects.filter(project=project, user=admin).first()
    if owner_membership is None:
        owner_role = project.roles.filter(slug="product-owner").first() or project.roles.first()
        owner_membership = Membership.objects.create(project=project, user=admin, role=owner_role, is_admin=True, email=admin.email)

    base_role = project.roles.filter(slug__in=["developer", "ux", "product-owner"]).first() or project.roles.first()
    base_permissions = list(base_role.permissions) if base_role else ["view_issues"]

    users_by_name = {}
    roles_by_department = {}
    employees = data.get("Employees", [])
    departments = sorted({text(row.get("Department")) for row in employees if text(row.get("Department"))})
    for order, department in enumerate(departments, start=100):
        role_slug = "department-" + slug(department, 36)
        role, _ = Role.objects.get_or_create(
            project=project,
            slug=role_slug,
            defaults={"name": department, "permissions": base_permissions, "order": order},
        )
        role.name = department
        role.permissions = base_permissions
        role.order = order
        role.save()
        roles_by_department[department] = role
    summary["department_roles"] = len(roles_by_department)

    for index, employee in enumerate(employees, start=1):
        full_name = text(employee.get("Full Name"))
        supplied_email = text(employee.get("Email")).lower()
        email = supplied_email if valid_email(supplied_email) else f"{slug(full_name) or f'employee-{index}'}@employees.invalid"
        username_base = slug(supplied_email.split("@", 1)[0] if valid_email(supplied_email) else full_name, 28) or f"employee-{index}"
        user = User.objects.filter(email__iexact=email).first()
        if user is None:
            username = username_base
            suffix = 2
            while User.objects.filter(username__iexact=username).exists():
                username = f"{username_base[:24]}-{suffix}"
                suffix += 1
            user = User(username=username, email=email)
            user.set_unusable_password()
            summary["users_created"] += 1
        else:
            summary["users_updated"] += 1
        user.full_name = full_name
        user.is_active = text(employee.get("Status")).lower() != "inactive"
        profile_lines = [
            f"Employee ID: {text(employee.get('Employee ID')) or 'Not assigned'}",
            f"Department: {text(employee.get('Department')) or 'Not assigned'}",
            f"Title: {text(employee.get('Title')) or 'Not assigned'}",
            f"Employee type: {text(employee.get('Employee Type')) or 'Not assigned'}",
            f"Employee status: {text(employee.get('Status')) or 'Not specified'}",
        ]
        if text(employee.get("Notes")):
            profile_lines.append(f"Notes: {text(employee.get('Notes'))}")
        user.bio = "\n".join(profile_lines)
        user.verified_email = valid_email(supplied_email)
        user.save()
        users_by_name[full_name.lower()] = user

        department = text(employee.get("Department"))
        role = roles_by_department.get(department) or base_role
        if role and user.is_active:
            membership, _ = Membership.objects.update_or_create(
                project=project,
                user=user,
                defaults={"role": role, "is_admin": False, "email": user.email, "invited_by": admin},
            )
            summary["memberships"] += 1

    type_names = ["IT Ticket", "Printer Asset", "Equipment Provision"]
    issue_types = {}
    for order, name in enumerate(type_names, start=100):
        obj, _ = IssueType.objects.get_or_create(project=project, name=name, defaults={"order": order, "color": "#4C566A"})
        issue_types[name] = obj

    status_labels = sorted({text(row.get("Status")) for sheet in ("Tickets", "Archive") for row in data.get(sheet, []) if text(row.get("Status"))})
    statuses = {}
    for order, label in enumerate(status_labels, start=100):
        lower = label.lower()
        closed = "complete" in lower or "cancel" in lower
        obj, _ = IssueStatus.objects.get_or_create(
            project=project,
            name=label,
            defaults={"order": order, "color": "#5E81AC", "is_closed": closed},
        )
        obj.is_closed = closed
        obj.save()
        statuses[label] = obj

    priority_labels = sorted({normalized_choice(row.get("Priority")) for sheet in ("Tickets", "Archive") for row in data.get(sheet, [])})
    priorities = {}
    priority_order = {"Critical": 10, "High": 20, "Medium": 30, "Normal": 40, "Low": 50, "Unspecified": 60}
    for label in priority_labels:
        obj, _ = Priority.objects.get_or_create(
            project=project,
            name=label,
            defaults={"order": priority_order.get(label, 100), "color": "#A3BE8C"},
        )
        priorities[label] = obj

    default_status = next(iter(statuses.values()), project.default_issue_status)
    default_priority = priorities.get("Normal") or next(iter(priorities.values()), project.default_priority)
    default_severity = project.default_severity or project.severities.first()

    attribute_names = [
        "Source Record Key", "Source Sheet", "External Record ID", "Original Type", "Original Priority",
        "Original Status", "Requester", "Assigned To", "Parent Task", "Start Date", "Due Date",
        "Completed Date", "Hours Estimated", "Hours Actual", "Imported Notes", "Location", "Department",
        "Model", "Primary User", "Asset Status", "Toner Model", "Toner Stock", "Supplier 1 Link",
        "Supplier 2 Link", "Provision Employee", "Provision Item Type", "Item Description", "Serial Number",
        "Asset Tag", "Issued Date", "Return Date", "Provision Status", "Condition",
    ]
    attributes = {}
    for order, name in enumerate(attribute_names, start=1):
        attribute, _ = IssueCustomAttribute.objects.get_or_create(
            project=project,
            name=name,
            defaults={"description": f"Imported workbook field: {name}", "type": "text", "order": order},
        )
        attributes[name] = attribute
    summary["custom_attributes"] = len(attributes)

    def resolve_person(name):
        name = text(name)
        if not name:
            return None
        user = users_by_name.get(name.lower())
        if user is None:
            summary["unresolved_people"] += 1
        return user

    def import_issue(sheet, record, entity_type):
        id_column = {"IT Ticket": "Ticket #", "Printer Asset": "Printer ID", "Equipment Provision": "Provision ID"}[entity_type]
        external_id = text(record.get(id_column))
        key = f"{sheet}:{external_id}"
        issue = Issue.objects.filter(project=project, external_reference__contains=[key]).first()
        created = issue is None
        if created:
            issue = Issue(project=project, external_reference=[key])
            summary["issues_created"] += 1
        else:
            summary["issues_updated"] += 1

        if entity_type == "IT Ticket":
            subject = text(record.get("Title")) or external_id
            requester = resolve_person(record.get("Requester"))
            assigned = resolve_person(record.get("Assigned To"))
            description_parts = [text(record.get("Description"))]
            if text(record.get("Notes")):
                description_parts.append("Imported notes:\n" + text(record.get("Notes")))
            status = statuses.get(text(record.get("Status")), default_status)
            priority = priorities.get(normalized_choice(record.get("Priority")), default_priority)
            tags = ["it-ticket", slug(normalized_choice(record.get("Type")))]
            due_date = parse_date(record.get("Due Date"))
            finished_date = aware_datetime(record.get("Completed Date"))
        elif entity_type == "Printer Asset":
            subject = " — ".join(filter(None, [external_id, text(record.get("Model")), text(record.get("Location"))]))
            requester = admin
            assigned = resolve_person(record.get("Primary User"))
            description_parts = [text(record.get("Notes"))]
            status = default_status
            priority = default_priority
            tags = ["printer-asset", slug(record.get("Status"))]
            due_date = None
            finished_date = None
        else:
            subject = " — ".join(filter(None, [external_id, text(record.get("Employee")), text(record.get("Item Description"))]))
            requester = admin
            assigned = resolve_person(record.get("Employee"))
            description_parts = [text(record.get("Notes"))]
            status = default_status
            priority = default_priority
            tags = ["equipment-provision", slug(record.get("Status"))]
            due_date = parse_date(record.get("Return Date"))
            finished_date = aware_datetime(record.get("Return Date"))

        issue.subject = f"[{external_id}] {subject}"[:500]
        issue.description = "\n\n".join(part for part in description_parts if part)
        issue.owner = requester or admin
        issue.assigned_to = assigned if assigned and assigned.is_active else None
        issue.type = issue_types[entity_type]
        issue.status = status
        issue.priority = priority
        issue.severity = default_severity
        issue.tags = [tag for tag in tags if tag]
        issue.due_date = due_date
        issue.finished_date = finished_date
        issue.save()

        values = {
            "Source Record Key": key,
            "Source Sheet": sheet,
            "External Record ID": external_id,
            "Original Type": record.get("Type"),
            "Original Priority": record.get("Priority"),
            "Original Status": record.get("Status"),
            "Requester": record.get("Requester"),
            "Assigned To": record.get("Assigned To"),
            "Parent Task": record.get("Parent Task"),
            "Start Date": record.get("Start Date"),
            "Due Date": record.get("Due Date"),
            "Completed Date": record.get("Completed Date"),
            "Hours Estimated": record.get("Hours Estimated"),
            "Hours Actual": record.get("Hours Actual"),
            "Imported Notes": record.get("Notes"),
            "Location": record.get("Location"),
            "Department": record.get("Department"),
            "Model": record.get("Model"),
            "Primary User": record.get("Primary User"),
            "Asset Status": record.get("Status") if entity_type == "Printer Asset" else "",
            "Toner Model": record.get("Toner Model"),
            "Toner Stock": record.get("Toner Stock"),
            "Supplier 1 Link": record.get("Supplier 1 Link"),
            "Supplier 2 Link": record.get("Supplier 2 Link"),
            "Provision Employee": record.get("Employee"),
            "Provision Item Type": record.get("Item Type"),
            "Item Description": record.get("Item Description"),
            "Serial Number": record.get("Serial Number"),
            "Asset Tag": record.get("Asset Tag"),
            "Issued Date": record.get("Issued Date"),
            "Return Date": record.get("Return Date"),
            "Provision Status": record.get("Status") if entity_type == "Equipment Provision" else "",
            "Condition": record.get("Condition"),
        }
        serialized = {str(attributes[name].id): text(value) for name, value in values.items() if text(value)}
        IssueCustomAttributesValues.objects.update_or_create(issue=issue, defaults={"attributes_values": serialized})

    for sheet in ("Tickets", "Archive"):
        for record in data.get(sheet, []):
            import_issue(sheet, record, "IT Ticket")
    for record in data.get("Printers", []):
        import_issue("Printers", record, "Printer Asset")
    for record in data.get("Provisions", []):
        import_issue("Provisions", record, "Equipment Provision")

    summary["project_id"] = project.id
    summary["issues_total"] = Issue.objects.filter(project=project).count()
    summary["active_members"] = Membership.objects.filter(project=project, user__is_active=True).count()

print(json.dumps(summary, sort_keys=True))
