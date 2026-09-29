# DefaultApi

All URIs are relative to *http://127.0.0.1:5000*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**changeMyPassword**](#changemypassword) | **PUT** /api/me/password | Change the logged-in user\&#39;s password|
|[**completeMyTodo**](#completemytodo) | **POST** /api/me/todos/{id}/complete | Complete a todo|
|[**createCategory**](#createcategory) | **POST** /api/categories | Add a task category (admin)|
|[**createMyTodo**](#createmytodo) | **POST** /api/me/todos | Add a todo|
|[**createMyWorkRecord**](#createmyworkrecord) | **POST** /api/me/work-records | Add a work record|
|[**createProject**](#createproject) | **POST** /api/projects | Add a project (admin)|
|[**createUser**](#createuser) | **POST** /api/users | Create a user (admin)|
|[**deleteCategory**](#deletecategory) | **DELETE** /api/categories/{id} | Delete a task category (admin)|
|[**deleteHoliday**](#deleteholiday) | **DELETE** /api/holidays/{date} | Delete a manual holiday entry (admin)|
|[**deleteMyTodo**](#deletemytodo) | **DELETE** /api/me/todos/{id} | Delete one of my todos|
|[**deleteMyWorkRecord**](#deletemyworkrecord) | **DELETE** /api/me/work-records/{id} | Delete one of my work records|
|[**deleteProject**](#deleteproject) | **DELETE** /api/projects/{id} | Delete a project (admin)|
|[**deleteUser**](#deleteuser) | **DELETE** /api/users/{account} | Delete a user (admin)|
|[**getMe**](#getme) | **GET** /api/me | Get the logged-in user|
|[**getWeekSummary**](#getweeksummary) | **GET** /api/me/week-summary | Summarize my week|
|[**getWorkOptions**](#getworkoptions) | **GET** /api/work/options | Get task categories, projects and the work table setting|
|[**listAllWorkRecords**](#listallworkrecords) | **GET** /api/work-records | List everyone\&#39;s work records|
|[**listHolidays**](#listholidays) | **GET** /api/holidays | List a year\&#39;s holiday calendar|
|[**listMissingEntries**](#listmissingentries) | **GET** /api/work/missing | List who missed logging work|
|[**listMyTodos**](#listmytodos) | **GET** /api/me/todos | List my todos|
|[**listMyWorkRecords**](#listmyworkrecords) | **GET** /api/me/work-records | List my work records|
|[**listUsers**](#listusers) | **GET** /api/users | List users (admin)|
|[**listWorkMembers**](#listworkmembers) | **GET** /api/work/members | List members for the everyone\&#39;s work table|
|[**login**](#login) | **POST** /api/login | Login|
|[**logout**](#logout) | **POST** /api/logout | Logout|
|[**saveHoliday**](#saveholiday) | **PUT** /api/holidays/{date} | Set a manual holiday or makeup workday (admin)|
|[**syncHolidays**](#syncholidays) | **POST** /api/holidays/sync | Sync the government office calendar now (admin)|
|[**updateCategory**](#updatecategory) | **PUT** /api/categories/{id} | Rename or (de)activate a task category (admin)|
|[**updateMe**](#updateme) | **PUT** /api/me | Update the logged-in user\&#39;s preferences|
|[**updateMyTodo**](#updatemytodo) | **PUT** /api/me/todos/{id} | Replace one of my todos|
|[**updateMyWorkRecord**](#updatemyworkrecord) | **PUT** /api/me/work-records/{id} | Replace one of my work records|
|[**updateProject**](#updateproject) | **PUT** /api/projects/{id} | Rename or (de)activate a project (admin)|
|[**updateUser**](#updateuser) | **PUT** /api/users/{account} | Update a user (admin)|
|[**updateWorkSetting**](#updateworksetting) | **PUT** /api/settings/work | Update the work table setting (admin)|

# **changeMyPassword**
> MessageResponse changeMyPassword(changeMyPasswordRequest)

Not allowed for the system (config) admin.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    ChangeMyPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let changeMyPasswordRequest: ChangeMyPasswordRequest; //

const { status, data } = await apiInstance.changeMyPassword(
    changeMyPasswordRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **changeMyPasswordRequest** | **ChangeMyPasswordRequest**|  | |


### Return type

**MessageResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **completeMyTodo**
> WorkRecordResponse completeMyTodo(completeTodoRequest)

Deletes the todo and adds it as a work record dated `date` (the client\'s today). Answers 400 and keeps the todo when the record would lack a category or hours.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    CompleteTodoRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)
let completeTodoRequest: CompleteTodoRequest; //

const { status, data } = await apiInstance.completeMyTodo(
    id,
    completeTodoRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **completeTodoRequest** | **CompleteTodoRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**WorkRecordResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createCategory**
> WorkOptionResponse createCategory(createWorkOptionRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    CreateWorkOptionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let createWorkOptionRequest: CreateWorkOptionRequest; //

const { status, data } = await apiInstance.createCategory(
    createWorkOptionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWorkOptionRequest** | **CreateWorkOptionRequest**|  | |


### Return type

**WorkOptionResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**409** | Conflict |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createMyTodo**
> TodoResponse createMyTodo(saveWorkEntryRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    SaveWorkEntryRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let saveWorkEntryRequest: SaveWorkEntryRequest; //

const { status, data } = await apiInstance.createMyTodo(
    saveWorkEntryRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **saveWorkEntryRequest** | **SaveWorkEntryRequest**|  | |


### Return type

**TodoResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createMyWorkRecord**
> WorkRecordResponse createMyWorkRecord(saveWorkEntryRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    SaveWorkEntryRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let saveWorkEntryRequest: SaveWorkEntryRequest; //

const { status, data } = await apiInstance.createMyWorkRecord(
    saveWorkEntryRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **saveWorkEntryRequest** | **SaveWorkEntryRequest**|  | |


### Return type

**WorkRecordResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createProject**
> WorkOptionResponse createProject(createWorkOptionRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    CreateWorkOptionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let createWorkOptionRequest: CreateWorkOptionRequest; //

const { status, data } = await apiInstance.createProject(
    createWorkOptionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWorkOptionRequest** | **CreateWorkOptionRequest**|  | |


### Return type

**WorkOptionResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**409** | Conflict |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createUser**
> UserResponse createUser(createUserRequest)

The account is stored in upper case. The new user\'s initial password is its account; passwords are case-insensitive.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    CreateUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let createUserRequest: CreateUserRequest; //

const { status, data } = await apiInstance.createUser(
    createUserRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createUserRequest** | **CreateUserRequest**|  | |


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**409** | Conflict |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteCategory**
> DeleteWorkOptionResponse deleteCategory()

Work records and todos using it get an empty categoryId. `cleared` is how many were changed.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteCategory(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**DeleteWorkOptionResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteHoliday**
> MessageResponse deleteHoliday()

The government entry of the date, if any, applies again.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let date: string; // (default to undefined)

const { status, data } = await apiInstance.deleteHoliday(
    date
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **date** | [**string**] |  | defaults to undefined|


### Return type

**MessageResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteMyTodo**
> MessageResponse deleteMyTodo()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteMyTodo(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**MessageResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteMyWorkRecord**
> MessageResponse deleteMyWorkRecord()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteMyWorkRecord(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**MessageResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteProject**
> DeleteWorkOptionResponse deleteProject()

Work records and todos using it get an empty projectId. `cleared` is how many were changed.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteProject(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**DeleteWorkOptionResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteUser**
> MessageResponse deleteUser()

The system admin and the caller themself cannot be deleted.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let account: string; // (default to undefined)

const { status, data } = await apiInstance.deleteUser(
    account
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **account** | [**string**] |  | defaults to undefined|


### Return type

**MessageResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMe**
> UserResponse getMe()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.getMe();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getWeekSummary**
> WeekSummaryResponse getWeekSummary()

Records and hours of the 7 days starting at `from`, and the hours to log: 8 per workday, where workdays are Monday-Friday minus holidays plus makeup workdays from the holiday calendar.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let from: string; //First day of the week (a Monday in the UI), YYYY-MM-DD. (default to undefined)

const { status, data } = await apiInstance.getWeekSummary(
    from
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **from** | [**string**] | First day of the week (a Monday in the UI), YYYY-MM-DD. | defaults to undefined|


### Return type

**WeekSummaryResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getWorkOptions**
> WorkOptionsResponse getWorkOptions()

Categories and projects include inactive ones, so old entries can still show their names.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.getWorkOptions();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**WorkOptionsResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listAllWorkRecords**
> WorkRecordListResponse listAllWorkRecords()

Every record dated from..to (the frontend asks for one week), newest date first. Allowed for admins, and for everyone when `allowViewAll` is on.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let from: string; //Inclusive start date (YYYY-MM-DD). (default to undefined)
let to: string; //Inclusive end date (YYYY-MM-DD), not before `from`. (default to undefined)
let account: string; // (optional) (default to undefined)
let categoryId: string; // (optional) (default to undefined)
let projectId: string; // (optional) (default to undefined)

const { status, data } = await apiInstance.listAllWorkRecords(
    from,
    to,
    account,
    categoryId,
    projectId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **from** | [**string**] | Inclusive start date (YYYY-MM-DD). | defaults to undefined|
| **to** | [**string**] | Inclusive end date (YYYY-MM-DD), not before &#x60;from&#x60;. | defaults to undefined|
| **account** | [**string**] |  | (optional) defaults to undefined|
| **categoryId** | [**string**] |  | (optional) defaults to undefined|
| **projectId** | [**string**] |  | (optional) defaults to undefined|


### Return type

**WorkRecordListResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listHolidays**
> HolidaysResponse listHolidays()

The effective entry per date (a manual entry wins over the government one).

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let year: string; // (default to undefined)

const { status, data } = await apiInstance.listHolidays(
    year
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **year** | [**string**] |  | defaults to undefined|


### Return type

**HolidaysResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listMissingEntries**
> MissingEntriesResponse listMissingEntries()

For everyone but the system admin, the workdays of the 30 days ending at `to` (but not before the work start date) without any work record (todos do not count). Weekends and holidays are skipped, makeup workdays count, and days before an account was created are skipped. Only members with missing days are listed, most missing first. Allowed for admins, and for everyone when `allowViewAll` is on.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let to: string; //Last day checked (the client\'s yesterday), YYYY-MM-DD. (default to undefined)

const { status, data } = await apiInstance.listMissingEntries(
    to
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **to** | [**string**] | Last day checked (the client\&#39;s yesterday), YYYY-MM-DD. | defaults to undefined|


### Return type

**MissingEntriesResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listMyTodos**
> TodosResponse listMyTodos()

Sorted by date, earliest first.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.listMyTodos();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**TodosResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listMyWorkRecords**
> WorkRecordListResponse listMyWorkRecords()

Every record dated from..to (the frontend asks for one week), newest date first.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let from: string; //Inclusive start date (YYYY-MM-DD). (default to undefined)
let to: string; //Inclusive end date (YYYY-MM-DD), not before `from`. (default to undefined)

const { status, data } = await apiInstance.listMyWorkRecords(
    from,
    to
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **from** | [**string**] | Inclusive start date (YYYY-MM-DD). | defaults to undefined|
| **to** | [**string**] | Inclusive end date (YYYY-MM-DD), not before &#x60;from&#x60;. | defaults to undefined|


### Return type

**WorkRecordListResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listUsers**
> ListUsersResponse listUsers()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.listUsers();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ListUsersResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWorkMembers**
> WorkMembersResponse listWorkMembers()

Allowed for admins, and for everyone when `allowViewAll` is on.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.listWorkMembers();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**WorkMembersResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> LoginResponse login(loginRequest)

Account and password are case-insensitive (accounts are stored in upper case). The returned JWT carries `sub` (account), `name`, `role` and `i18n` claims.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let loginRequest: LoginRequest; //

const { status, data } = await apiInstance.login(
    loginRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **loginRequest** | **LoginRequest**|  | |


### Return type

**LoginResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **logout**
> logout()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.logout();
```

### Parameters
This endpoint does not have any parameters.


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | No Content |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **saveHoliday**
> HolidayResponse saveHoliday(saveHolidayRequest)

Creates or replaces the manual entry of the date; it wins over the government entry.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    SaveHolidayRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let date: string; // (default to undefined)
let saveHolidayRequest: SaveHolidayRequest; //

const { status, data } = await apiInstance.saveHoliday(
    date,
    saveHolidayRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **saveHolidayRequest** | **SaveHolidayRequest**|  | |
| **date** | [**string**] |  | defaults to undefined|


### Return type

**HolidayResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **syncHolidays**
> SyncHolidaysResponse syncHolidays()

Replaces the government entries of this year and the next; manual entries are kept.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.syncHolidays();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**SyncHolidaysResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**502** | The office calendar could not be fetched |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateCategory**
> WorkOptionResponse updateCategory(updateWorkOptionRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    UpdateWorkOptionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)
let updateWorkOptionRequest: UpdateWorkOptionRequest; //

const { status, data } = await apiInstance.updateCategory(
    id,
    updateWorkOptionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateWorkOptionRequest** | **UpdateWorkOptionRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**WorkOptionResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateMe**
> TokenResponse updateMe(updateMeRequest)

Returns a new JWT carrying the updated `i18n` claim.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    UpdateMeRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let updateMeRequest: UpdateMeRequest; //

const { status, data } = await apiInstance.updateMe(
    updateMeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateMeRequest** | **UpdateMeRequest**|  | |


### Return type

**TokenResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateMyTodo**
> TodoResponse updateMyTodo(saveWorkEntryRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    SaveWorkEntryRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)
let saveWorkEntryRequest: SaveWorkEntryRequest; //

const { status, data } = await apiInstance.updateMyTodo(
    id,
    saveWorkEntryRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **saveWorkEntryRequest** | **SaveWorkEntryRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**TodoResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateMyWorkRecord**
> WorkRecordResponse updateMyWorkRecord(saveWorkEntryRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    SaveWorkEntryRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)
let saveWorkEntryRequest: SaveWorkEntryRequest; //

const { status, data } = await apiInstance.updateMyWorkRecord(
    id,
    saveWorkEntryRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **saveWorkEntryRequest** | **SaveWorkEntryRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**WorkRecordResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateProject**
> WorkOptionResponse updateProject(updateWorkOptionRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    UpdateWorkOptionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; // (default to undefined)
let updateWorkOptionRequest: UpdateWorkOptionRequest; //

const { status, data } = await apiInstance.updateProject(
    id,
    updateWorkOptionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateWorkOptionRequest** | **UpdateWorkOptionRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**WorkOptionResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUser**
> UserResponse updateUser(updateUserRequest)

Only the given fields are changed. The system admin cannot be updated.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    UpdateUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let account: string; // (default to undefined)
let updateUserRequest: UpdateUserRequest; //

const { status, data } = await apiInstance.updateUser(
    account,
    updateUserRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateUserRequest** | **UpdateUserRequest**|  | |
| **account** | [**string**] |  | defaults to undefined|


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateWorkSetting**
> WorkSettingResponse updateWorkSetting(updateWorkSettingRequest)


### Example

```typescript
import {
    DefaultApi,
    Configuration,
    UpdateWorkSettingRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let updateWorkSettingRequest: UpdateWorkSettingRequest; //

const { status, data } = await apiInstance.updateWorkSetting(
    updateWorkSettingRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateWorkSettingRequest** | **UpdateWorkSettingRequest**|  | |


### Return type

**WorkSettingResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

