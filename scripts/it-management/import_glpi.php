<?php

declare(strict_types=1);

use Glpi\Kernel\Kernel;

if ($argc !== 2) {
    fwrite(STDERR, "usage: php import_glpi.php NORMALIZED_JSON\n");
    exit(2);
}

$sourcePath = $argv[1];
if (!is_file($sourcePath)) {
    throw new RuntimeException("normalized JSON not found: {$sourcePath}");
}

require '/var/www/glpi/vendor/autoload.php';

$kernel = new Kernel();
$kernel->boot();

global $DB;

$adminLogin = getenv('GLPI_IMPORT_ADMIN_LOGIN') ?: 'glpi';
$adminPassword = getenv('GLPI_IMPORT_ADMIN_PASSWORD') ?: '';
if ($adminPassword === '') {
    throw new RuntimeException('GLPI_IMPORT_ADMIN_PASSWORD is required');
}

$auth = new Auth();
if (!$auth->login($adminLogin, $adminPassword, true)) {
    throw new RuntimeException('GLPI administrator authentication failed');
}

$data = json_decode((string) file_get_contents($sourcePath), true, 512, JSON_THROW_ON_ERROR);
$summary = array_fill_keys([
    'users_created', 'users_updated', 'groups_created', 'groups_updated',
    'printers_created', 'printers_updated', 'provisions_created', 'provisions_updated',
    'tickets_created', 'tickets_updated', 'unresolved_people',
], 0);

$text = static fn(mixed $value): string => trim((string) ($value ?? ''));
$slug = static function (string $value): string {
    $value = strtolower($value);
    $value = preg_replace('/[^a-z0-9]+/', '-', $value) ?? '';
    return trim(substr($value, 0, 48), '-');
};
$splitName = static function (string $fullName): array {
    $parts = preg_split('/\s+/', trim($fullName), -1, PREG_SPLIT_NO_EMPTY) ?: [];
    return [$parts[0] ?? 'Employee', count($parts) > 1 ? implode(' ', array_slice($parts, 1)) : 'Record'];
};
$parseDate = static function (mixed $value) use ($text): ?string {
    $value = $text($value);
    if ($value === '') {
        return null;
    }
    try {
        return (new DateTimeImmutable($value))->format('Y-m-d H:i:s');
    } catch (Throwable) {
        return null;
    }
};
$validEmail = static fn(string $value): bool => filter_var($value, FILTER_VALIDATE_EMAIL) !== false;
$sourceKey = static function (string $prefix, mixed $id, array $row): string {
    $id = trim((string) ($id ?? ''));
    if ($id !== '') {
        return "ITMS-{$prefix}:{$id}";
    }
    return "ITMS-{$prefix}-HASH:" . substr(hash('sha256', json_encode($row, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES)), 0, 20);
};
$fail = static function (string $model, array $input): never {
    $errors = $_SESSION['MESSAGE_AFTER_REDIRECT'] ?? [];
    throw new RuntimeException("failed to save {$model}: " . json_encode($errors, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));
};

$upsertModel = static function (CommonDBTM $model, array $criteria, array $input) use ($fail): array {
    if ($model->getFromDBByCrit($criteria)) {
        $id = (int) $model->getID();
        if (!$model->update(['id' => $id] + $input)) {
            $fail($model::class, $input);
        }
        return [$id, false];
    }
    $id = (int) $model->add($input);
    if ($id <= 0) {
        $fail($model::class, $input);
    }
    return [$id, true];
};

$dropdown = static function (string $class, string $name, array $extra = []) use ($upsertModel): int {
    if ($name === '') {
        return 0;
    }
    /** @var CommonDBTM $model */
    $model = new $class();
    [$id] = $upsertModel($model, ['name' => $name], ['name' => $name] + $extra);
    return $id;
};

$findUserByName = [];
$employeeRows = $data['Employees'] ?? [];

$DB->beginTransaction();
try {
    $selfServiceProfile = new Profile();
    if (!$selfServiceProfile->getFromDBByCrit(['name' => 'Self-Service'])) {
        throw new RuntimeException('GLPI Self-Service profile not found');
    }
    $selfServiceProfileId = (int) $selfServiceProfile->getID();

    foreach ($employeeRows as $index => $employee) {
        $fullName = $text($employee['Full Name'] ?? '');
        [$firstname, $realname] = $splitName($fullName);
        $key = $sourceKey('EMP', $employee['Employee ID'] ?? '', $employee);
        $loginBase = $slug($fullName);
        if ($loginBase === '') {
            $loginBase = 'employee-' . ($index + 1);
        }
        $loginBase = 'itms-' . $loginBase;

        $user = new User();
        $exists = $user->getFromDBByCrit(['registration_number' => $key]);
        if ($exists) {
            $login = (string) $user->fields['name'];
        } else {
            $login = $loginBase;
            $suffix = 2;
            $probe = new User();
            while ($probe->getFromDBByCrit(['name' => $login])) {
                $login = $loginBase . '-' . $suffix++;
                $probe = new User();
            }
        }

        $comment = implode("\n", array_filter([
            "Imported from IT_Management_System.xlsm ({$key})",
            'Department: ' . $text($employee['Department'] ?? ''),
            'Job title: ' . $text($employee['Title'] ?? ''),
            'Employee type: ' . $text($employee['Employee Type'] ?? ''),
            'Access Control ID: ' . $text($employee['Access Control ID'] ?? ''),
            'Time Clock ID: ' . $text($employee['Time Clock ID'] ?? ''),
            'Time Clock Card ID: ' . $text($employee['Time Clock Card ID'] ?? ''),
            'Source status: ' . $text($employee['Status'] ?? ''),
            $text($employee['Notes'] ?? ''),
        ], static fn(string $line): bool => !str_ends_with($line, ': ') && $line !== ''));

        $input = [
            'name' => $login,
            'firstname' => substr($firstname, 0, 255),
            'realname' => substr($realname, 0, 255),
            'registration_number' => $key,
            'comment' => $comment,
            'is_active' => strcasecmp($text($employee['Status'] ?? ''), 'Inactive') === 0 ? 0 : 1,
            'entities_id' => 0,
        ];
        if (!$exists) {
            $password = 'Itm!' . bin2hex(random_bytes(24));
            $input['password'] = $password;
            $input['password2'] = $password;
        }
        [$userId, $created] = $upsertModel($user, ['registration_number' => $key], $input);
        $summary[$created ? 'users_created' : 'users_updated']++;

        $suppliedEmail = strtolower($text($employee['Email'] ?? ''));
        $email = $validEmail($suppliedEmail)
            ? $suppliedEmail
            : ($slug($fullName) ?: 'employee-' . ($index + 1)) . '@employees.invalid';
        $userEmail = new UserEmail();
        $upsertModel($userEmail, ['users_id' => $userId, 'is_default' => 1], [
            'users_id' => $userId,
            'email' => $email,
            'is_default' => 1,
            'is_dynamic' => 0,
        ]);

        $profileUser = new Profile_User();
        $upsertModel($profileUser, [
            'users_id' => $userId,
            'profiles_id' => $selfServiceProfileId,
            'entities_id' => 0,
        ], [
            'users_id' => $userId,
            'profiles_id' => $selfServiceProfileId,
            'entities_id' => 0,
            'is_recursive' => 1,
            'is_dynamic' => 0,
        ]);

        $department = $text($employee['Department'] ?? '');
        if ($department !== '') {
            $groupName = 'ITMS: ' . $department;
            $group = new Group();
            [$groupId, $groupCreated] = $upsertModel($group, ['name' => $groupName], [
                'name' => $groupName,
                'entities_id' => 0,
                'is_recursive' => 1,
                'is_usergroup' => 1,
                'comment' => 'Imported department from IT_Management_System.xlsm',
            ]);
            $summary[$groupCreated ? 'groups_created' : 'groups_updated']++;
            $membership = new Group_User();
            $upsertModel($membership, ['groups_id' => $groupId, 'users_id' => $userId], [
                'groups_id' => $groupId,
                'users_id' => $userId,
                'is_dynamic' => 0,
                'is_manager' => 0,
                'is_userdelegate' => 0,
            ]);
        }

        if ($fullName !== '') {
            $findUserByName[strtolower($fullName)] = $userId;
        }
    }

    foreach (($data['Printers'] ?? []) as $printerRow) {
        $key = $sourceKey('PRINTER', $printerRow['Printer ID'] ?? '', $printerRow);
        $modelName = $text($printerRow['Model'] ?? '');
        $locationName = $text($printerRow['Location'] ?? '');
        $primaryUserName = $text($printerRow['Primary User'] ?? '');
        $primaryUserId = $findUserByName[strtolower($primaryUserName)] ?? 0;
        if ($primaryUserName !== '' && $primaryUserName !== '-' && $primaryUserId === 0) {
            $summary['unresolved_people']++;
        }
        $comment = implode("\n", array_filter([
            "Imported from IT_Management_System.xlsm ({$key})",
            'Department: ' . $text($printerRow['Department'] ?? ''),
            'Primary user: ' . $primaryUserName,
            'Source status: ' . $text($printerRow['Status'] ?? ''),
            'Toner model: ' . $text($printerRow['Toner Model'] ?? ''),
            'Toner stock: ' . $text($printerRow['Toner Stock'] ?? ''),
            'Supplier 1: ' . $text($printerRow['Supplier 1 Link'] ?? ''),
            'Supplier 2: ' . $text($printerRow['Supplier 2 Link'] ?? ''),
            $text($printerRow['Notes'] ?? ''),
        ], static fn(string $line): bool => !str_ends_with($line, ': ') && $line !== ''));
        $printer = new Printer();
        [$printerId, $created] = $upsertModel($printer, ['otherserial' => $key], [
            'name' => substr(implode(' — ', array_filter([$text($printerRow['Printer ID'] ?? ''), $modelName, $locationName])), 0, 255),
            'otherserial' => $key,
            'entities_id' => 0,
            'is_recursive' => 0,
            'printermodels_id' => $dropdown(PrinterModel::class, $modelName),
            'locations_id' => $dropdown(Location::class, $locationName, ['entities_id' => 0]),
            'states_id' => $dropdown(State::class, $text($printerRow['Status'] ?? '')),
            'users_id' => $primaryUserId,
            'contact' => $primaryUserName,
            'comment' => $comment,
        ]);
        $summary[$created ? 'printers_created' : 'printers_updated']++;
    }

    foreach (($data['Provisions'] ?? []) as $provisionRow) {
        $key = $sourceKey('PROVISION', $provisionRow['Provision ID'] ?? '', $provisionRow);
        $employeeName = $text($provisionRow['Employee'] ?? '');
        $userId = $findUserByName[strtolower($employeeName)] ?? 0;
        if ($employeeName !== '' && $userId === 0) {
            $summary['unresolved_people']++;
        }
        $itemType = $text($provisionRow['Item Type'] ?? '');
        $description = $text($provisionRow['Item Description'] ?? '');
        $assetTag = $text($provisionRow['Asset Tag'] ?? '');
        $comment = implode("\n", array_filter([
            "Imported from IT_Management_System.xlsm ({$key})",
            'Employee: ' . $employeeName,
            'Issued date: ' . $text($provisionRow['Issued Date'] ?? ''),
            'Return date: ' . $text($provisionRow['Return Date'] ?? ''),
            'Source status: ' . $text($provisionRow['Status'] ?? ''),
            'Condition: ' . $text($provisionRow['Condition'] ?? ''),
            $text($provisionRow['Notes'] ?? ''),
        ], static fn(string $line): bool => !str_ends_with($line, ': ') && $line !== ''));
        $peripheral = new Peripheral();
        [$peripheralId, $created] = $upsertModel($peripheral, ['otherserial' => $key], [
            'name' => substr(implode(' — ', array_filter([$text($provisionRow['Provision ID'] ?? ''), $itemType, $description])), 0, 255),
            'otherserial' => $key,
            'serial' => $text($provisionRow['Serial Number'] ?? ''),
            'entities_id' => 0,
            'is_recursive' => 0,
            'peripheraltypes_id' => $dropdown(PeripheralType::class, $itemType),
            'states_id' => $dropdown(State::class, $text($provisionRow['Status'] ?? '')),
            'users_id' => $userId,
            'contact' => $employeeName,
            'contact_num' => $assetTag,
            'comment' => $comment,
        ]);
        $summary[$created ? 'provisions_created' : 'provisions_updated']++;
    }

    $statusValue = static function (string $value): int {
        $value = strtolower($value);
        return match (true) {
            str_contains($value, 'complete') => Ticket::CLOSED,
            str_contains($value, 'cancel') => Ticket::CLOSED,
            str_contains($value, 'ongoing') => Ticket::ASSIGNED,
            str_contains($value, 'back'), str_contains($value, 'incomplete') => Ticket::WAITING,
            default => Ticket::INCOMING,
        };
    };
    $priorityValue = static function (string $value): int {
        $value = strtolower($value);
        return match (true) {
            str_contains($value, 'critical') => 6,
            str_contains($value, 'high') => 4,
            str_contains($value, 'low') => 2,
            default => 3,
        };
    };

    foreach (['Tickets', 'Archive'] as $sheet) {
        foreach (($data[$sheet] ?? []) as $ticketRow) {
            $key = $sourceKey(strtoupper($sheet), $ticketRow['Ticket #'] ?? '', $ticketRow);
            $requesterName = $text($ticketRow['Requester'] ?? '');
            $assigneeName = $text($ticketRow['Assigned To'] ?? '');
            $requesterId = $findUserByName[strtolower($requesterName)] ?? 0;
            $assigneeId = $findUserByName[strtolower($assigneeName)] ?? 0;
            if ($requesterName !== '' && $requesterId === 0) {
                $summary['unresolved_people']++;
            }
            if ($assigneeName !== '' && $assigneeId === 0) {
                $summary['unresolved_people']++;
            }
            $title = $text($ticketRow['Title'] ?? '');
            $status = $statusValue($text($ticketRow['Status'] ?? ''));
            $completedDate = $parseDate($ticketRow['Completed Date'] ?? '');
            $content = implode("\n\n", array_filter([
                "Source key: {$key}",
                $text($ticketRow['Description'] ?? ''),
                'Original type: ' . $text($ticketRow['Type'] ?? ''),
                'Original priority: ' . $text($ticketRow['Priority'] ?? ''),
                'Original status: ' . $text($ticketRow['Status'] ?? ''),
                'Original requester: ' . $requesterName,
                'Parent task reference: ' . $text($ticketRow['Parent Task'] ?? ''),
                'Hours estimated: ' . $text($ticketRow['Hours Estimated'] ?? ''),
                'Hours actual: ' . $text($ticketRow['Hours Actual'] ?? ''),
                $text($ticketRow['Notes'] ?? '') === '' ? '' : "Imported notes:\n" . $text($ticketRow['Notes'] ?? ''),
            ], static fn(string $part): bool => !str_ends_with($part, ': ') && $part !== ''));

            $ticket = new Ticket();
            $exists = $ticket->getFromDBByCrit(['externalid' => $key]);
            $input = [
                'externalid' => $key,
                'name' => substr(($title === '' ? $key : $title), 0, 255),
                'content' => $content,
                'entities_id' => 0,
                'type' => Ticket::DEMAND_TYPE,
                'status' => $status,
                'priority' => $priorityValue($text($ticketRow['Priority'] ?? '')),
                'urgency' => $priorityValue($text($ticketRow['Priority'] ?? '')),
                'impact' => 3,
                'itilcategories_id' => $dropdown(ITILCategory::class, 'ITMS: ' . ($text($ticketRow['Type'] ?? '') ?: 'Uncategorized'), ['entities_id' => 0]),
                'users_id_recipient' => $requesterId ?: Session::getLoginUserID(),
                '_users_id_requester' => $requesterId ?: Session::getLoginUserID(),
                '_users_id_assign' => $assigneeId ?: 0,
            ];
            $startDate = $parseDate($ticketRow['Start Date'] ?? '');
            $dueDate = $parseDate($ticketRow['Due Date'] ?? '');
            if ($startDate !== null) {
                $input['date'] = $startDate;
            }
            if ($dueDate !== null) {
                $input['time_to_resolve'] = $dueDate;
            }
            if ($completedDate !== null && in_array($status, [Ticket::SOLVED, Ticket::CLOSED], true)) {
                $input['solvedate'] = $completedDate;
                $input['closedate'] = $completedDate;
            }
            [$ticketId, $created] = $upsertModel($ticket, ['externalid' => $key], $input);
            $summary[$created ? 'tickets_created' : 'tickets_updated']++;
        }
    }

    $DB->commit();
} catch (Throwable $error) {
    $DB->rollback();
    throw $error;
}

ksort($summary);
echo json_encode($summary, JSON_THROW_ON_ERROR | JSON_UNESCAPED_SLASHES) . PHP_EOL;
