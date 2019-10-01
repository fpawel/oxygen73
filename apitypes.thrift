// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

struct Product {
  1: i32 place,
  2: i64 productID,
  3: i64 partyID,
  4: i32 serial,
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
}

struct Bucket{
    1: i64 bucketID
    2: TimeUnixMillis createdAt
    3: TimeUnixMillis updatedAt
    4: i64 partyID
    5: TimeUnixMillis partyCreatedAt
    6: bool isLast
}

struct YearMonth{
    1: i32 year
    2: i32 month
}

struct Measurement{
    1: TimeUnixMillis StoredAt
    2: double temperature
    3: double pressure
    4: double humidity
    5: list<double> places
}

struct LogEntry{
    1: TimeUnixMillis time
    2: string line
}