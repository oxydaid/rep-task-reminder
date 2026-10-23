module.exports = {
  apps: [
    {
      name: 'whatsapp-reminder',
      script: './bin/whatsapp-reminder',
      cwd: __dirname,
      instances: 1, // Must be 1 to prevent scheduler conflict or WhatsApp session conflict
      autorestart: true,
      watch: false,
      max_memory_restart: '300M',
      min_uptime: '10s',
      max_restarts: 10,
      restart_delay: 3000,
      kill_timeout: 10000,
      env: {
        NODE_ENV: 'production',
        PORT: '8080',
        DB_PATH: './data/reminders.db',
        WA_DB_PATH: './data/whatsapp.db',
        BACKUP_DIR: './data/backups',
        STATIC_DIR: './frontend/dist',
        DEFAULT_TIMEZONE: 'Asia/Jakarta',
      },
      error_file: './data/logs/pm2-error.log',
      out_file: './data/logs/pm2-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss',
    },
  ],
}
