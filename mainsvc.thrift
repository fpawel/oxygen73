
include "apitypes.thrift"

service MainSvc {
   apitypes.Party getParty(1:i64 partyID)
   list<apitypes.Product> listProducts(1:i64 partyID)
   list<apitypes.YearMonth> listYearMonths()
   list<apitypes.Bucket> listBucketsOfYearMonth(1:i32 year, 2:i32 month)
   oneway void requestMeasurements(1:apitypes.TimeUnixMillis timeFrom, 2:apitypes.TimeUnixMillis timeTo)
   oneway void createNewParty()

   list<apitypes.Product> listLastPartyProducts()
   oneway void setLastPartyProductSerialAtPlace(1:i32 place, 2:i32 serial)

   list<apitypes.TimeUnixMillis> ListLogEntriesDays()
   list<apitypes.LogEntry> LogEntriesOfDay(1:apitypes.TimeUnixMillis daytime, 2:string filter)
}