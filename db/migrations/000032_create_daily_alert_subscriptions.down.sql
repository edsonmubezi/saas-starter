DROP TABLE IF EXISTS daily_alert_subscriptions;

DELETE FROM permissions WHERE name = 'tenant.daily_alert.manage';
