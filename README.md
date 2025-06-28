# Pocket ID Analytics

A lightweight analytics service that collects heartbeat data from Pocket ID instances to count active deployments.

## Overview

Seeing how many active Pocket ID instances are out there through our analytics server genuinely motivates our team to keep developing and maintaining the project. The instance count is also displayed on the [Pocket ID website](https://pocket-id.org).

## Data Collection

Only minimal, non-identifiable data is collected, and analytics can be completely disabled by users.

The server stores only the following information:

| Field            | Description                                                |
| :--------------- | :--------------------------------------------------------- |
| **Instance ID**  | A unique, non-identifiable UUID for the Pocket ID instance |
| **First seen**   | Timestamp when the instance first sent a heartbeat         |
| **Last seen**    | Timestamp of the most recent heartbeat                     |
| **Last version** | Version of the Pocket ID instance                          |

### Activity Status

- **Active**: Instance has sent a heartbeat within the last day after initial registration
- **Inactive**: No heartbeat received for 2+ consecutive days

## API Endpoints

This server is hosted at `https://analytics.pocket-id.org`.

### Get Statistics

```http
GET /stats
```

Returns active instance count and historical data.

**Query Parameters:**

- `timeframe` (optional): Data timeframe
  - `daily` - Daily counts for last 30 days (default)
  - `monthly` - Monthly counts

**Example Response:**

```json
{
  "total": 5,
  "history": [
    {
      "date": "2025-05-23",
      "count": 1
    },
    {
      "date": "2025-05-24",
      "count": 3
    },
    {
      "date": "2025-05-25",
      "count": 5
    }
  ]
}
```

### Send Heartbeat

```http
POST /heartbeat
```

Registers or updates an instance's heartbeat.

**Request Body:**

```json
{
  "instance_id": "b316815f-5f81-488f-89f8-12b62013dfa4",
  "version": "1.0.0"
}
```

**Parameters:**

- `instance_id` (string, required): Unique UUID for the Pocket ID instance
- `version` (string, required): Current version of the Pocket ID instance

---

_For more information about Pocket ID, visit the [main repository](https://github.com/pocket-id/pocket-id)._
