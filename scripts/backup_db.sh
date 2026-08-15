#!/bin/bash
# =============================================================================
# AUTOMATED BACKUP SCRIPT FOR POSTGRESQL & MINIO STORAGE (PESAN ANTAR DESA)
# =============================================================================

BACKUP_DIR="./backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
DB_BACKUP_FILE="${BACKUP_DIR}/postgres_backup_${TIMESTAMP}.sql"

mkdir -p ${BACKUP_DIR}

echo "📦 [Backup Script] Starting PostgreSQL Database Backup..."
docker exec desa_postgres pg_dump -U desa_user order_db > ${DB_BACKUP_FILE}

if [ $? -eq 0 ]; then
    echo "✅ [Backup Script] PostgreSQL Backup Success: ${DB_BACKUP_FILE}"
else
    echo "❌ [Backup Script] PostgreSQL Backup Failed!"
    exit 1
fi

echo "📦 [Backup Script] Database Backup Complete!"
