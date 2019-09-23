
include "apitypes.thrift"

service MainSvc {
   apitypes.Party getParty(1:i64 partyID)
   list<apitypes.Product> listProducts(1:i64 partyID)
   list<apitypes.YearMonth> listYearMonths()
   list<apitypes.Bucket> listBucketsOfYearMonth(1:i32 year, 2:i32 month)
   list<apitypes.Measurement> listMeasurements(1:apitypes.TimeUnixMillis timeFrom, 2:apitypes.TimeUnixMillis timeTo)
   oneway void createNewParty(1:list<apitypes.Product> products)
   oneway void openClient()
   oneway void closeClient()
}