<?php
/**
 * Job Queue System for Long-Running Operations
 * SQLite-based job tracking with simple background execution
 */

const JOBS_DB_PATH = '/tmp/devarch-jobs.db';

/**
 * Initialize job database
 */
function initJobsDatabase(): void {
    $db = new PDO('sqlite:' . JOBS_DB_PATH);
    $db->exec('
        CREATE TABLE IF NOT EXISTS jobs (
            id TEXT PRIMARY KEY,
            type TEXT NOT NULL,
            status TEXT NOT NULL,
            progress_current INTEGER DEFAULT 0,
            progress_total INTEGER DEFAULT 100,
            current_task TEXT,
            result TEXT,
            error TEXT,
            logs TEXT,
            created_at INTEGER NOT NULL,
            updated_at INTEGER NOT NULL,
            completed_at INTEGER
        )
    ');

    // Create index for faster queries
    $db->exec('CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)');
    $db->exec('CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(type)');
}

/**
 * Create new job
 */
function createJob(string $type, array $data = []): string {
    initJobsDatabase();
    $db = new PDO('sqlite:' . JOBS_DB_PATH);

    $id = uniqid($type . '-', true);
    $stmt = $db->prepare('
        INSERT INTO jobs (id, type, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)
    ');

    $now = time();
    $stmt->execute([$id, $type, 'pending', $now, $now]);

    return $id;
}

/**
 * Update job progress
 */
function updateJob(string $id, string $status, ?array $progress = null, ?array $result = null, ?string $error = null, ?array $logs = null): bool {
    initJobsDatabase();
    $db = new PDO('sqlite:' . JOBS_DB_PATH);

    $updates = ['status = ?', 'updated_at = ?'];
    $params = [$status, time()];

    if ($progress !== null) {
        if (isset($progress['current'])) {
            $updates[] = 'progress_current = ?';
            $params[] = $progress['current'];
        }
        if (isset($progress['total'])) {
            $updates[] = 'progress_total = ?';
            $params[] = $progress['total'];
        }
        if (isset($progress['currentTask'])) {
            $updates[] = 'current_task = ?';
            $params[] = $progress['currentTask'];
        }
    }

    if ($result !== null) {
        $updates[] = 'result = ?';
        $params[] = json_encode($result);
    }

    if ($error !== null) {
        $updates[] = 'error = ?';
        $params[] = $error;
    }

    if ($logs !== null) {
        $updates[] = 'logs = ?';
        $params[] = json_encode($logs);
    }

    if ($status === 'completed' || $status === 'failed') {
        $updates[] = 'completed_at = ?';
        $params[] = time();
    }

    $sql = 'UPDATE jobs SET ' . implode(', ', $updates) . ' WHERE id = ?';
    $params[] = $id;

    $stmt = $db->prepare($sql);
    return $stmt->execute($params);
}

/**
 * Get job details
 */
function getJob(string $id): ?array {
    initJobsDatabase();
    $db = new PDO('sqlite:' . JOBS_DB_PATH);

    $stmt = $db->prepare('SELECT * FROM jobs WHERE id = ?');
    $stmt->execute([$id]);
    $job = $stmt->fetch(PDO::FETCH_ASSOC);

    if (!$job) {
        return null;
    }

    // Parse JSON fields
    if ($job['result']) {
        $job['result'] = json_decode($job['result'], true);
    }
    if ($job['logs']) {
        $job['logs'] = json_decode($job['logs'], true);
    }

    // Calculate duration
    if ($job['completed_at']) {
        $job['duration'] = $job['completed_at'] - $job['created_at'];
    } elseif ($job['status'] === 'in_progress') {
        $job['duration'] = time() - $job['created_at'];
    } else {
        $job['duration'] = null;
    }

    // Calculate percentage
    if ($job['progress_total'] > 0) {
        $job['progress_percentage'] = round(($job['progress_current'] / $job['progress_total']) * 100, 1);
    } else {
        $job['progress_percentage'] = 0;
    }

    return $job;
}

/**
 * List jobs with filters
 */
function listJobs(?string $status = null, ?string $type = null, int $limit = 50): array {
    initJobsDatabase();
    $db = new PDO('sqlite:' . JOBS_DB_PATH);

    $where = [];
    $params = [];

    if ($status !== null) {
        $where[] = 'status = ?';
        $params[] = $status;
    }

    if ($type !== null) {
        $where[] = 'type = ?';
        $params[] = $type;
    }

    $sql = 'SELECT * FROM jobs';
    if (!empty($where)) {
        $sql .= ' WHERE ' . implode(' AND ', $where);
    }
    $sql .= ' ORDER BY created_at DESC LIMIT ?';
    $params[] = $limit;

    $stmt = $db->prepare($sql);
    $stmt->execute($params);

    $jobs = $stmt->fetchAll(PDO::FETCH_ASSOC);

    // Parse JSON fields and calculate durations
    foreach ($jobs as &$job) {
        if ($job['result']) {
            $job['result'] = json_decode($job['result'], true);
        }
        if ($job['logs']) {
            $job['logs'] = json_decode($job['logs'], true);
        }

        if ($job['completed_at']) {
            $job['duration'] = $job['completed_at'] - $job['created_at'];
        } elseif ($job['status'] === 'in_progress') {
            $job['duration'] = time() - $job['created_at'];
        } else {
            $job['duration'] = null;
        }

        if ($job['progress_total'] > 0) {
            $job['progress_percentage'] = round(($job['progress_current'] / $job['progress_total']) * 100, 1);
        } else {
            $job['progress_percentage'] = 0;
        }
    }

    return $jobs;
}

/**
 * Delete job
 */
function deleteJob(string $id): bool {
    initJobsDatabase();
    $db = new PDO('sqlite:' . JOBS_DB_PATH);

    // Check if job is running
    $job = getJob($id);
    if (!$job) {
        return false;
    }

    if ($job['status'] === 'in_progress') {
        // Mark as cancelled instead of deleting
        return updateJob($id, 'cancelled');
    }

    // Delete completed/failed jobs
    $stmt = $db->prepare('DELETE FROM jobs WHERE id = ?');
    return $stmt->execute([$id]);
}

/**
 * Clean up old jobs (older than specified days)
 */
function cleanupOldJobs(int $daysOld = 7): int {
    initJobsDatabase();
    $db = new PDO('sqlite:' . JOBS_DB_PATH);

    $cutoff = time() - ($daysOld * 86400);
    $stmt = $db->prepare('DELETE FROM jobs WHERE completed_at < ? OR (status = ? AND created_at < ?)');
    $stmt->execute([$cutoff, 'failed', $cutoff]);

    return $stmt->rowCount();
}

/**
 * Execute job in background
 * Uses exec with & to run in background (simple approach)
 */
function executeJobInBackground(string $jobId, string $workerScript): bool {
    $cmd = sprintf(
        'php %s %s > /dev/null 2>&1 &',
        escapeshellarg($workerScript),
        escapeshellarg($jobId)
    );

    exec($cmd, $output, $returnCode);

    return $returnCode === 0;
}

/**
 * Add log entry to job
 */
function addJobLog(string $jobId, string $message): void {
    $job = getJob($jobId);
    if (!$job) {
        return;
    }

    $logs = $job['logs'] ?? [];
    $logs[] = [
        'timestamp' => time(),
        'message' => $message
    ];

    updateJob($jobId, $job['status'], null, null, null, $logs);
}
