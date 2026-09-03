# frozen_string_literal: true

require "json"
require "securerandom"
require "digest"
require "date"

abort "usage: rails runner import_openproject.rb NORMALIZED_JSON" unless ARGV[0]
data = JSON.parse(File.read(ARGV[0]))

PROJECT_IDENTIFIER = "it-management-system"
PROJECT_NAME = "IT Management System"
TYPE_NAMES = {
  ticket: "IT Ticket",
  printer: "Printer Asset",
  provision: "Equipment Provision"
}.freeze

slugify = lambda do |value|
  value.to_s.downcase.gsub(/[^a-z0-9]+/, "-").gsub(/^-|-$/, "")[0, 48]
end

as_text = ->(value) { value.nil? ? "" : value.to_s.strip }
parse_date = lambda do |value|
  text = as_text.call(value)
  text.empty? ? nil : Date.iso8601(text)
rescue Date::Error
  nil
end

split_name = lambda do |full_name|
  parts = as_text.call(full_name).split
  [parts.first || "Employee", parts.length > 1 ? parts[1..].join(" ") : "Record"]
end

valid_email = ->(value) { as_text.call(value).match?(/\A[^@\s]+@[^@\s]+\.[^@\s]+\z/) }

summary = Hash.new(0)
unresolved_people = []
unresolved_parents = []

ActiveRecord::Base.transaction do
  admin = User.admin.active.first || abort("OpenProject has no active administrator")
  project = Project.find_or_initialize_by(identifier: PROJECT_IDENTIFIER)
  project.assign_attributes(name: PROJECT_NAME, public: false, workspace_type: "project", description: "Imported from IT_Management_System.xlsm")
  project.save!

  requested_modules = %w[work_package_tracking calendar_view costs]
  available_modules = OpenProject::AccessControl.available_project_modules.map(&:to_s)
  project.enabled_module_names = requested_modules & available_modules
  project.save!

  types = TYPE_NAMES.transform_values do |name|
    Type.find_or_create_by!(name: name) do |type|
      type.position = Type.maximum(:position).to_i + 1
      type.is_default = false
      type.is_milestone = false
      type.is_standard = true if type.respond_to?(:is_standard=)
    end
  end
  project.types = (project.types.to_a + types.values).uniq

  role = Role.find_by(type: "ProjectRole", name: "Member") || Role.where(type: "ProjectRole").where.not(id: [1, 2]).first
  abort "OpenProject has no assignable project role" unless role

  user_field_names = [
    "Employee ID", "Department", "Job Title", "Employee Type", "Access Control ID",
    "Time Clock ID", "Time Clock Card ID", "Employee Status", "Employee Notes"
  ]
  user_fields = user_field_names.to_h do |name|
    field = UserCustomField.find_or_initialize_by(name: name)
    field.field_format = "string" if field.new_record?
    field.assign_attributes(editable: true, is_required: false, searchable: true,
                            custom_field_section: UserCustomFieldSection.first)
    field.save!
    [name, field]
  end

  field_specs = {
    "Source Record Key" => "string",
    "Source Sheet" => "string",
    "External Record ID" => "string",
    "Original Type" => "string",
    "Original Priority" => "string",
    "Original Status" => "string",
    "Original Requester" => "string",
    "Imported Start Date" => "date",
    "Imported Due Date" => "date",
    "Completed Date" => "date",
    "Hours Estimated (Imported)" => "float",
    "Hours Actual (Imported)" => "float",
    "Imported Notes" => "text",
    "Parent Task Reference" => "string",
    "Location" => "string",
    "Department" => "string",
    "Model" => "string",
    "Primary User" => "string",
    "Asset Status" => "string",
    "Toner Model" => "string",
    "Toner Stock" => "int",
    "Supplier 1 Link" => "string",
    "Supplier 2 Link" => "string",
    "Provision Employee" => "string",
    "Provision Item Type" => "string",
    "Item Description" => "text",
    "Serial Number" => "string",
    "Asset Tag" => "string",
    "Issued Date" => "date",
    "Return Date" => "date",
    "Provision Status" => "string",
    "Condition" => "string"
  }
  fields = field_specs.to_h do |name, format|
    field = WorkPackageCustomField.find_or_initialize_by(name: name)
    field.field_format = format if field.new_record?
    field.assign_attributes(is_for_all: false, editable: true, is_required: false, searchable: %w[string text].include?(format))
    field.save!
    field.projects << project unless field.projects.exists?(project.id)
    types.values.each { |type| field.types << type unless field.types.exists?(type.id) }
    [name, field]
  end

  set_custom_values = lambda do |record, values|
    values.each do |name, value|
      next if value.nil? || as_text.call(value).empty?
      field = fields[name] || user_fields[name]
      next unless field
      custom_value = CustomValue.find_or_initialize_by(customized: record, custom_field: field)
      custom_value.value = value
      custom_value.save!
    end
  end

  users_by_name = {}
  data.fetch("Employees", []).each_with_index do |employee, index|
    full_name = as_text.call(employee["Full Name"])
    firstname, lastname = split_name.call(full_name)
    supplied_email = as_text.call(employee["Email"]).downcase
    email = if valid_email.call(supplied_email)
              supplied_email
            else
              base = slugify.call(full_name)
              "#{base.empty? ? "employee-#{index + 1}" : base}@employees.invalid"
            end
    login_base = slugify.call(valid_email.call(supplied_email) ? supplied_email.split("@").first : full_name)
    login_base = "employee-#{index + 1}" if login_base.empty?
    user = User.where("LOWER(mail) = ?", email.downcase).first
    if user.nil?
      login = login_base
      suffix = 2
      while User.where("LOWER(login) = ?", login.downcase).exists?
        login = "#{login_base}-#{suffix}"
        suffix += 1
      end
      user = User.new(login: login, mail: email, language: "en")
      password = "Itm!#{SecureRandom.hex(24)}"
      user.password = password
      user.password_confirmation = password
      user.force_password_change = true
      summary[:users_created] += 1
    else
      summary[:users_updated] += 1
    end
    user.firstname = firstname[0, 30]
    user.lastname = lastname[0, 30]
    user.status = as_text.call(employee["Status"]).casecmp("Inactive").zero? ? "locked" : "active"
    user.save!

    set_custom_values.call(user, {
      "Employee ID" => employee["Employee ID"],
      "Department" => employee["Department"],
      "Job Title" => employee["Title"],
      "Employee Type" => employee["Employee Type"],
      "Access Control ID" => employee["Access Control ID"],
      "Time Clock ID" => employee["Time Clock ID"],
      "Time Clock Card ID" => employee["Time Clock Card ID"],
      "Employee Status" => employee["Status"],
      "Employee Notes" => employee["Notes"]
    })

    department = as_text.call(employee["Department"])
    unless department.empty?
      group = Group.find_or_create_by!(lastname: department)
      GroupUser.find_or_create_by!(group: group, user: user)
    end
    member = Member.find_or_initialize_by(project: project, principal: user)
    member.roles = [role]
    member.save!
    users_by_name[full_name.downcase] = user
  end

  status_map = lambda do |label|
    text = as_text.call(label).downcase
    target = if text.include?("complete") then "Closed"
             elsif text.include?("cancel") then "Rejected"
             elsif text.include?("ongoing") then "In progress"
             elsif text.include?("back") then "On hold"
             else "New"
             end
    Status.find_by(name: target) || Status.first
  end
  priority_map = lambda do |label|
    text = as_text.call(label).downcase
    target = if text.include?("critical") then "Immediate"
             elsif text.include?("high") then "High"
             elsif text.include?("low") then "Low"
             else "Normal"
             end
    IssuePriority.find_by(name: target) || IssuePriority.first
  end

  work_packages_by_key = {}
  ticket_keys_by_sheet_and_id = {}

  import_record = lambda do |sheet, record, kind|
    id_column = { ticket: "Ticket #", printer: "Printer ID", provision: "Provision ID" }.fetch(kind)
    external_id = as_text.call(record[id_column])
    key = "#{sheet}:#{external_id}"
    existing_value = CustomValue.find_by(custom_field: fields["Source Record Key"], value: key)
    work_package = existing_value&.customized_type == "WorkPackage" ? WorkPackage.find_by(id: existing_value.customized_id) : nil
    if work_package
      summary[:work_packages_updated] += 1
    else
      work_package = WorkPackage.new(project: project)
      summary[:work_packages_created] += 1
    end

    title = kind == :ticket ? as_text.call(record["Title"]) : if kind == :printer
      [external_id, as_text.call(record["Model"]), as_text.call(record["Location"])].reject(&:empty?).join(" — ")
    else
      [external_id, as_text.call(record["Employee"]), as_text.call(record["Item Description"])].reject(&:empty?).join(" — ")
    end
    title = external_id if title.empty?
    requester_name = kind == :ticket ? record["Requester"] : nil
    assignee_name = case kind
                    when :ticket then record["Assigned To"]
                    when :printer then record["Primary User"]
                    when :provision then record["Employee"]
                    end
    requester = users_by_name[as_text.call(requester_name).downcase]
    assignee = users_by_name[as_text.call(assignee_name).downcase]
    unresolved_people << "#{key}:requester" if !as_text.call(requester_name).empty? && requester.nil?
    unresolved_people << "#{key}:assignee" if !as_text.call(assignee_name).empty? && assignee.nil?

    description_parts = []
    description_parts << as_text.call(record["Description"]) unless as_text.call(record["Description"]).empty?
    description_parts << "Imported notes:\n#{as_text.call(record["Notes"])}" unless as_text.call(record["Notes"]).empty?
    work_package.assign_attributes(
      project: project,
      type: types.fetch(kind),
      subject: title[0, 255],
      description: description_parts.join("\n\n"),
      author: requester || admin,
      assigned_to: assignee,
      status: kind == :ticket ? status_map.call(record["Status"]) : Status.find_by(name: "New") || Status.first,
      priority: kind == :ticket ? priority_map.call(record["Priority"]) : IssuePriority.find_by(name: "Normal") || IssuePriority.first
    )
    work_package.save!

    common_values = {
      "Source Record Key" => key,
      "Source Sheet" => sheet,
      "External Record ID" => external_id,
      "Original Type" => record["Type"],
      "Original Priority" => record["Priority"],
      "Original Status" => record["Status"],
      "Original Requester" => record["Requester"],
      "Imported Start Date" => parse_date.call(record["Start Date"]),
      "Imported Due Date" => parse_date.call(record["Due Date"]),
      "Completed Date" => parse_date.call(record["Completed Date"]),
      "Hours Estimated (Imported)" => record["Hours Estimated"],
      "Hours Actual (Imported)" => record["Hours Actual"],
      "Imported Notes" => record["Notes"],
      "Parent Task Reference" => record["Parent Task"]
    }
    kind_values = case kind
                  when :printer
                    {
                      "Location" => record["Location"], "Department" => record["Department"],
                      "Model" => record["Model"], "Primary User" => record["Primary User"],
                      "Asset Status" => record["Status"], "Toner Model" => record["Toner Model"],
                      "Toner Stock" => record["Toner Stock"], "Supplier 1 Link" => record["Supplier 1 Link"],
                      "Supplier 2 Link" => record["Supplier 2 Link"]
                    }
                  when :provision
                    {
                      "Provision Employee" => record["Employee"], "Provision Item Type" => record["Item Type"],
                      "Item Description" => record["Item Description"], "Serial Number" => record["Serial Number"],
                      "Asset Tag" => record["Asset Tag"], "Issued Date" => parse_date.call(record["Issued Date"]),
                      "Return Date" => parse_date.call(record["Return Date"]), "Provision Status" => record["Status"],
                      "Condition" => record["Condition"]
                    }
                  else
                    {}
                  end
    set_custom_values.call(work_package, common_values.merge(kind_values))
    work_packages_by_key[key] = work_package
    ticket_keys_by_sheet_and_id[[sheet, external_id]] = work_package if kind == :ticket
  end

  data.fetch("Tickets", []).each { |record| import_record.call("Tickets", record, :ticket) }
  data.fetch("Archive", []).each { |record| import_record.call("Archive", record, :ticket) }
  data.fetch("Printers", []).each { |record| import_record.call("Printers", record, :printer) }
  data.fetch("Provisions", []).each { |record| import_record.call("Provisions", record, :provision) }

  %w[Tickets Archive].each do |sheet|
    data.fetch(sheet, []).each do |record|
      child = ticket_keys_by_sheet_and_id[[sheet, as_text.call(record["Ticket #"])]]
      parent_ref = as_text.call(record["Parent Task"])
      next if parent_ref.empty? || child.nil?
      parent = ticket_keys_by_sheet_and_id[[sheet, parent_ref]] || ticket_keys_by_sheet_and_id[["Tickets", parent_ref]] || ticket_keys_by_sheet_and_id[["Archive", parent_ref]]
      if parent && parent != child
        child.update!(parent: parent)
        summary[:parent_links] += 1
      else
        unresolved_parents << "#{sheet}:#{as_text.call(record["Ticket #"])}=>#{parent_ref}"
      end
    end
  end

  summary[:project_id] = project.id
  summary[:project_members] = project.members.count
  summary[:groups] = Group.where(lastname: data.fetch("Employees", []).map { |row| as_text.call(row["Department"]) }.reject(&:empty?).uniq).count
  summary[:work_packages_total] = project.work_packages.count
  summary[:custom_fields] = fields.length + user_fields.length
  summary[:unresolved_people] = unresolved_people.length
  summary[:unresolved_parents] = unresolved_parents.length
end

puts JSON.generate(summary.sort.to_h)
