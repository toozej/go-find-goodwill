# go-find-goodwill Architecture Diagram

## System Overview

```mermaid
graph TD
    %% Main Components
    A[User] -->|CLI Commands| B[CLI Interface]
    A -->|Web Browser| C[Web UI]
    B --> D[Core Service]
    C --> D
    D --> E[ShopGoodwill API Client]
    D --> F[Database Layer]
    D --> G[Notification Service]
    D --> H[Scheduling System]

    %% Database Layer
    F --> I[SQLite Database]
    I --> J[Searches Table]
    I --> K[Items Table]
    I --> L[Search History]
    I --> M[Price History]
    I --> N[Notification Log]

    %% ShopGoodwill API Client
    E --> O[Authentication]
    E --> P[Search Execution]
    E --> Q[Item Details]
    E --> R[Anti-Bot Measures]

    %% Notification Service
    G --> S[Gotify Client]
    G --> T[Email Notifier]
    G --> U[Notification Queue]

    %% Anti-Bot Measures
    R --> V[User Agent Rotation]
    R --> W[Exponential Backoff]
    R --> X[Randomized Timing]

    %% Scheduling System
    H --> Y[Cron Scheduler]
    H --> Z[Interval Manager]

    %% Data Flow
    D -->|Read/Write| F
    E -->|Fetch Data| D
    D -->|Send Alerts| G
    G -->|Deliver| S
    G -->|Deliver| T

    %% Configuration
    A -->|Environment Vars| D
    A -->|CLI Flags| D
```

## Component Interaction Flow

### Search Execution Flow
```mermaid
sequenceDiagram
    participant User
    participant Scheduler
    participant SearchManager
    participant APIClient
    participant Database
    participant Deduplication
    participant Notification

    User->>Scheduler: Configure search interval
    loop Every X minutes
        Scheduler->>SearchManager: Trigger search execution
        SearchManager->>Database: Get active searches
        Database-->>SearchManager: Return search queries
        loop For each search
            SearchManager->>APIClient: Execute search(query)
            APIClient->>APIClient: Apply anti-bot measures
            APIClient->>ShopGoodwill: HTTP Request
            ShopGoodwill-->>APIClient: Search results
            APIClient-->>SearchManager: Processed results
            SearchManager->>Deduplication: Check for duplicates
            Deduplication->>Database: Query existing items
            Database-->>Deduplication: Existing items
            Deduplication-->>SearchManager: New items only
            SearchManager->>Database: Store new items
            SearchManager->>Database: Update search history
            SearchManager->>Notification: Queue notifications
        end
        Notification->>Database: Get notification preferences
        Database-->>Notification: User preferences
        loop For each notification
            Notification->>Gotify: Send alert
            Notification->>Email: Send email
        end
    end
```

## Detailed Component Architecture

### 1. CLI Interface Architecture
```mermaid
classDiagram
    class CLI {
        +Execute()
        +rootCmd
        +init()
    }

    class RootCommand {
        +Run()
        +PersistentPreRun()
        +Flags
    }

    class SearchSubcommand {
        +Add()
        +Remove()
        +List()
        +Edit()
    }

    class ConfigSubcommand {
        +Show()
        +Set()
    }

    class DatabaseSubcommand {
        +Migrate()
        +Backup()
        +Restore()
    }

    CLI --> RootCommand
    RootCommand --> SearchSubcommand
    RootCommand --> ConfigSubcommand
    RootCommand --> DatabaseSubcommand
```

### 2. Core Service Architecture
```mermaid
classDiagram
    class CoreService {
        +Start()
        +Stop()
        +ProcessSearchResults()
    }

    class SearchManager {
        +ExecuteSearches()
        +AddSearch()
        +RemoveSearch()
        +UpdateSearch()
    }

    class ResultProcessor {
        +ProcessItems()
        +ExtractDetails()
        +UpdatePricing()
    }

    class DeduplicationEngine {
        +CheckDuplicate()
        +GetItemFingerprint()
        +UpdateItemTracking()
    }

    CoreService --> SearchManager
    CoreService --> ResultProcessor
    CoreService --> DeduplicationEngine
```

### 3. Database Layer Architecture
```mermaid
classDiagram
    class Database {
        <<interface>>
        +Connect()
        +Close()
        +Migrate()
    }

    class SQLiteRepository {
        +GetSearches()
        +AddSearch()
        +UpdateSearch()
        +DeleteSearch()
        +GetItems()
        +AddItem()
        +UpdateItem()
        +GetSearchHistory()
        +AddSearchExecution()
    }

    class MigrationManager {
        +RunMigrations()
        +CreateMigration()
        +GetCurrentVersion()
    }

    Database <|-- SQLiteRepository
    Database --> MigrationManager
```

### 4. Notification System Architecture
```mermaid
classDiagram
    class NotificationService {
        +SendNotification()
        +QueueNotification()
        +GetDeliveryStatus()
    }

    class Notifier {
        <<interface>>
        +Send()
        +ValidateConfig()
    }

    class GotifyNotifier {
        +Send()
        +CreateMessage()
    }

    class EmailNotifier {
        +Send()
        +FormatEmail()
    }

    class NotificationQueue {
        +AddToQueue()
        +ProcessQueue()
        +GetQueueStatus()
    }

    NotificationService --> NotificationQueue
    NotificationService --> Notifier
    Notifier <|-- GotifyNotifier
    Notifier <|-- EmailNotifier
```

### 5. Anti-Bot Measures Architecture
```mermaid
classDiagram
    class AntiBotManager {
        +ApplyMeasures()
        +RotateUserAgent()
        +CalculateDelay()
    }

    class UserAgentRotator {
        +GetRandomUserAgent()
        +AddUserAgent()
        +RemoveUserAgent()
    }

    class TimingManager {
        +CalculateJitter()
        +GetRandomizedInterval()
        +ApplyExponentialBackoff()
    }

    class RequestThrottler {
        +CheckRateLimit()
        +WaitForNextRequest()
        +ResetCounter()
    }

    AntiBotManager --> UserAgentRotator
    AntiBotManager --> TimingManager
    AntiBotManager --> RequestThrottler
```

## Data Flow Architecture

### Search Result Processing Flow
```mermaid
flowchart TD
    A[ShopGoodwill API] -->|Raw Results| B[API Client]
    B -->|Processed Results| C[Search Manager]
    C --> D{New Items?}
    D -->|Yes| E[Deduplication Check]
    D -->|No| F[Update Existing Items]
    E --> G{Duplicate Found?}
    G -->|No| H[Add New Item to DB]
    G -->|Yes| I[Update Item Metadata]
    H --> J[Queue Notification]
    F --> J
    I --> J
    J --> K[Notification Service]
    K --> L[Gotify Client]
    K --> M[Email Client]
```

### Database Interaction Flow
```mermaid
flowchart TD
    A[Core Service] -->|Search Request| B[Database]
    B --> C[Searches Table]
    C -->|Active Searches| D[Search Manager]
    D -->|Search Results| E[Items Table]
    E --> F{Deduplication Check}
    F -->|New Item| G[Add to Items]
    F -->|Existing Item| H[Update Item]
    G --> I[Search-Item Mapping]
    H --> I
    I --> J[Search History]
    J --> K[Update Last Checked]
```

## Deployment Architecture

### Containerized Deployment
```mermaid
graph TD
    A[User] -->|HTTP| B[Web Server:8080]
    A -->|CLI| C[Application Container]
    C --> D[SQLite Database Volume]
    C --> E[Configuration Volume]
    C --> F[Log Volume]
    B --> C
    C --> G[ShopGoodwill API]
    C --> H[Gotify Server]
    C --> I[SMTP Server]
```

### Configuration Management
```mermaid
graph TD
    A[Environment Variables] --> B[Config Parser]
    C[.env File] --> B
    D[CLI Flags] --> B
    B --> E[Configuration Struct]
    E --> F[Core Service]
    E --> G[Database]
    E --> H[API Client]
    E --> I[Notification Service]
```

## Error Handling Architecture

```mermaid
flowchart TD
    A[Operation] --> B{Success?}
    B -->|Yes| C[Continue]
    B -->|No| D[Error Handler]
    D --> E[Log Error]
    D --> F[Check Retry Policy]
    F --> G{Retriable?}
    G -->|Yes| H[Apply Backoff]
    H --> I[Retry Operation]
    G -->|No| J[Notify User]
    J --> K[Continue/Exit]
```

This comprehensive architecture diagram provides a visual representation of all major components, their interactions, and the data flow throughout the go-find-goodwill application.