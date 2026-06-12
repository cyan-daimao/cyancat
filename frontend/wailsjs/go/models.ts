export namespace api {
	
	export class Page__cyancat_internal_adapter_dto_ConnectionDTO_ {
	    page: number;
	    pageSize: number;
	    total: number;
	    list: dto.ConnectionDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Page__cyancat_internal_adapter_dto_ConnectionDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.total = source["total"];
	        this.list = this.convertValues(source["list"], dto.ConnectionDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Page__cyancat_internal_adapter_dto_QueryHistoryDTO_ {
	    page: number;
	    pageSize: number;
	    total: number;
	    list: dto.QueryHistoryDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Page__cyancat_internal_adapter_dto_QueryHistoryDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.total = source["total"];
	        this.list = this.convertValues(source["list"], dto.QueryHistoryDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response__cyancat_internal_adapter_dto_ConnectionDTO_ {
	    code: number;
	    message: string;
	    data?: dto.ConnectionDTO;
	
	    static createFrom(source: any = {}) {
	        return new Response__cyancat_internal_adapter_dto_ConnectionDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.ConnectionDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response__cyancat_internal_adapter_dto_QueryResultDTO_ {
	    code: number;
	    message: string;
	    data?: dto.QueryResultDTO;
	
	    static createFrom(source: any = {}) {
	        return new Response__cyancat_internal_adapter_dto_QueryResultDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.QueryResultDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response__cyancat_internal_adapter_dto_TableDetailDTO_ {
	    code: number;
	    message: string;
	    data?: dto.TableDetailDTO;
	
	    static createFrom(source: any = {}) {
	        return new Response__cyancat_internal_adapter_dto_TableDetailDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.TableDetailDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response__cyancat_internal_adapter_dto_TestConnectionResultDTO_ {
	    code: number;
	    message: string;
	    data?: dto.TestConnectionResultDTO;
	
	    static createFrom(source: any = {}) {
	        return new Response__cyancat_internal_adapter_dto_TestConnectionResultDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.TestConnectionResultDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response__cyancat_internal_infra_api_Page__cyancat_internal_adapter_dto_ConnectionDTO__ {
	    code: number;
	    message: string;
	    // Go type: Page__cyancat_internal_adapter_dto_ConnectionDTO_
	    data?: any;
	
	    static createFrom(source: any = {}) {
	        return new Response__cyancat_internal_infra_api_Page__cyancat_internal_adapter_dto_ConnectionDTO__(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response__cyancat_internal_infra_api_Page__cyancat_internal_adapter_dto_QueryHistoryDTO__ {
	    code: number;
	    message: string;
	    // Go type: Page__cyancat_internal_adapter_dto_QueryHistoryDTO_
	    data?: any;
	
	    static createFrom(source: any = {}) {
	        return new Response__cyancat_internal_infra_api_Page__cyancat_internal_adapter_dto_QueryHistoryDTO__(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response____cyancat_internal_adapter_dto_CharsetDTO_ {
	    code: number;
	    message: string;
	    data: dto.CharsetDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Response____cyancat_internal_adapter_dto_CharsetDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.CharsetDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response____cyancat_internal_adapter_dto_CollationDTO_ {
	    code: number;
	    message: string;
	    data: dto.CollationDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Response____cyancat_internal_adapter_dto_CollationDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.CollationDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response____cyancat_internal_adapter_dto_ConnectionDTO_ {
	    code: number;
	    message: string;
	    data: dto.ConnectionDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Response____cyancat_internal_adapter_dto_ConnectionDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.ConnectionDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response____cyancat_internal_adapter_dto_DatabaseDTO_ {
	    code: number;
	    message: string;
	    data: dto.DatabaseDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Response____cyancat_internal_adapter_dto_DatabaseDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.DatabaseDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response____cyancat_internal_adapter_dto_SchemaDTO_ {
	    code: number;
	    message: string;
	    data: dto.SchemaDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Response____cyancat_internal_adapter_dto_SchemaDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.SchemaDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response____cyancat_internal_adapter_dto_TableDTO_ {
	    code: number;
	    message: string;
	    data: dto.TableDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Response____cyancat_internal_adapter_dto_TableDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.TableDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response____cyancat_internal_adapter_dto_ViewDTO_ {
	    code: number;
	    message: string;
	    data: dto.ViewDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Response____cyancat_internal_adapter_dto_ViewDTO_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], dto.ViewDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Response_bool_ {
	    code: number;
	    message: string;
	    data: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Response_bool_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = source["data"];
	    }
	}
	export class Response_string_ {
	    code: number;
	    message: string;
	    data: string;
	
	    static createFrom(source: any = {}) {
	        return new Response_string_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = source["data"];
	    }
	}

}

export namespace dto {
	
	export class ForeignKeySpecDTO {
	    name: string;
	    columns: string[];
	    referencedSchema: string;
	    referencedTable: string;
	    referencedColumns: string[];
	    onDelete: string;
	    onUpdate: string;
	
	    static createFrom(source: any = {}) {
	        return new ForeignKeySpecDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.columns = source["columns"];
	        this.referencedSchema = source["referencedSchema"];
	        this.referencedTable = source["referencedTable"];
	        this.referencedColumns = source["referencedColumns"];
	        this.onDelete = source["onDelete"];
	        this.onUpdate = source["onUpdate"];
	    }
	}
	export class IndexSpecDTO {
	    name: string;
	    type: string;
	    columns: string[];
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new IndexSpecDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.columns = source["columns"];
	        this.comment = source["comment"];
	    }
	}
	export class ColumnRenameDTO {
	    old: string;
	    new: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnRenameDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.old = source["old"];
	        this.new = source["new"];
	    }
	}
	export class ColumnSpecDTO {
	    name: string;
	    dataType: string;
	    typeLength?: number;
	    precision?: number;
	    scale?: number;
	    nullable: boolean;
	    autoIncrement: boolean;
	    unsigned: boolean;
	    defaultValue?: string;
	    comment: string;
	    collation: string;
	    first: boolean;
	    after: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnSpecDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.dataType = source["dataType"];
	        this.typeLength = source["typeLength"];
	        this.precision = source["precision"];
	        this.scale = source["scale"];
	        this.nullable = source["nullable"];
	        this.autoIncrement = source["autoIncrement"];
	        this.unsigned = source["unsigned"];
	        this.defaultValue = source["defaultValue"];
	        this.comment = source["comment"];
	        this.collation = source["collation"];
	        this.first = source["first"];
	        this.after = source["after"];
	    }
	}
	export class AlterTableRequest {
	    connID: number;
	    database: string;
	    schema: string;
	    name: string;
	    addColumns: ColumnSpecDTO[];
	    dropColumns: string[];
	    renameColumns: ColumnRenameDTO[];
	    modifyColumns: ColumnSpecDTO[];
	    addIndexes: IndexSpecDTO[];
	    dropIndexes: string[];
	    addForeignKeys: ForeignKeySpecDTO[];
	    dropForeignKeys: string[];
	    engine: string;
	    charset: string;
	    collation: string;
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new AlterTableRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.database = source["database"];
	        this.schema = source["schema"];
	        this.name = source["name"];
	        this.addColumns = this.convertValues(source["addColumns"], ColumnSpecDTO);
	        this.dropColumns = source["dropColumns"];
	        this.renameColumns = this.convertValues(source["renameColumns"], ColumnRenameDTO);
	        this.modifyColumns = this.convertValues(source["modifyColumns"], ColumnSpecDTO);
	        this.addIndexes = this.convertValues(source["addIndexes"], IndexSpecDTO);
	        this.dropIndexes = source["dropIndexes"];
	        this.addForeignKeys = this.convertValues(source["addForeignKeys"], ForeignKeySpecDTO);
	        this.dropForeignKeys = source["dropForeignKeys"];
	        this.engine = source["engine"];
	        this.charset = source["charset"];
	        this.collation = source["collation"];
	        this.comment = source["comment"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CharsetDTO {
	    name: string;
	    description: string;
	    defaultCollation: string;
	
	    static createFrom(source: any = {}) {
	        return new CharsetDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.defaultCollation = source["defaultCollation"];
	    }
	}
	export class CollationDTO {
	    name: string;
	    charset: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CollationDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.charset = source["charset"];
	        this.isDefault = source["isDefault"];
	    }
	}
	export class ColumnDTO {
	    name: string;
	    databaseType: string;
	    nullable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ColumnDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.databaseType = source["databaseType"];
	        this.nullable = source["nullable"];
	    }
	}
	
	
	export class ConnectionDTO {
	    id: number;
	    name: string;
	    type: string;
	    host: string;
	    port: number;
	    user: string;
	    database: string;
	    ssl: boolean;
	    group: string;
	    color: string;
	    // Go type: time
	    lastConnectedAt?: any;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.database = source["database"];
	        this.ssl = source["ssl"];
	        this.group = source["group"];
	        this.color = source["color"];
	        this.lastConnectedAt = this.convertValues(source["lastConnectedAt"], null);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateConnectionRequest {
	    name: string;
	    type: string;
	    host: string;
	    port: number;
	    user: string;
	    password: string;
	    database: string;
	    ssl: boolean;
	    group: string;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateConnectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.ssl = source["ssl"];
	        this.group = source["group"];
	        this.color = source["color"];
	    }
	}
	export class CreateDatabaseRequest {
	    connID: number;
	    name: string;
	    charset: string;
	    collation: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateDatabaseRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.name = source["name"];
	        this.charset = source["charset"];
	        this.collation = source["collation"];
	    }
	}
	export class CreateTableRequest {
	    connID: number;
	    database: string;
	    schema: string;
	    name: string;
	    columns: ColumnSpecDTO[];
	    primaryKey: string[];
	    indexes: IndexSpecDTO[];
	    foreignKeys: ForeignKeySpecDTO[];
	    engine: string;
	    charset: string;
	    collation: string;
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateTableRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.database = source["database"];
	        this.schema = source["schema"];
	        this.name = source["name"];
	        this.columns = this.convertValues(source["columns"], ColumnSpecDTO);
	        this.primaryKey = source["primaryKey"];
	        this.indexes = this.convertValues(source["indexes"], IndexSpecDTO);
	        this.foreignKeys = this.convertValues(source["foreignKeys"], ForeignKeySpecDTO);
	        this.engine = source["engine"];
	        this.charset = source["charset"];
	        this.collation = source["collation"];
	        this.comment = source["comment"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DatabaseDTO {
	    name: string;
	    charset: string;
	    collation: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.charset = source["charset"];
	        this.collation = source["collation"];
	    }
	}
	export class ExecuteQueryRequest {
	    connID: number;
	    sql: string;
	    stream: boolean;
	    maxRows: number;
	    database: string;
	    schema: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteQueryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.sql = source["sql"];
	        this.stream = source["stream"];
	        this.maxRows = source["maxRows"];
	        this.database = source["database"];
	        this.schema = source["schema"];
	    }
	}
	export class ForeignKeyDTO {
	    name: string;
	    columns: string[];
	    referencedSchema: string;
	    referencedTable: string;
	    referencedColumns: string[];
	    onUpdate: string;
	    onDelete: string;
	
	    static createFrom(source: any = {}) {
	        return new ForeignKeyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.columns = source["columns"];
	        this.referencedSchema = source["referencedSchema"];
	        this.referencedTable = source["referencedTable"];
	        this.referencedColumns = source["referencedColumns"];
	        this.onUpdate = source["onUpdate"];
	        this.onDelete = source["onDelete"];
	    }
	}
	
	export class GetCreateTableDDLRequest {
	    connID: number;
	    database: string;
	    schema: string;
	    table: string;
	
	    static createFrom(source: any = {}) {
	        return new GetCreateTableDDLRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.database = source["database"];
	        this.schema = source["schema"];
	        this.table = source["table"];
	    }
	}
	export class IndexDTO {
	    name: string;
	    columns: string[];
	    unique: boolean;
	    primary: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IndexDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.columns = source["columns"];
	        this.unique = source["unique"];
	        this.primary = source["primary"];
	    }
	}
	
	export class ListCharsetsRequest {
	    connID: number;
	
	    static createFrom(source: any = {}) {
	        return new ListCharsetsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	    }
	}
	export class ListCollationsRequest {
	    connID: number;
	    charset: string;
	
	    static createFrom(source: any = {}) {
	        return new ListCollationsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.charset = source["charset"];
	    }
	}
	export class ListConnectionRequest {
	    group: string;
	    type: string;
	    keyword: string;
	
	    static createFrom(source: any = {}) {
	        return new ListConnectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.group = source["group"];
	        this.type = source["type"];
	        this.keyword = source["keyword"];
	    }
	}
	export class PageConnectionRequest {
	    group: string;
	    type: string;
	    keyword: string;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new PageConnectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.group = source["group"];
	        this.type = source["type"];
	        this.keyword = source["keyword"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	export class QueryHistoryDTO {
	    id: number;
	    connID: number;
	    sql: string;
	    status: string;
	    errorMessage: string;
	    rowCount: number;
	    durationMs: number;
	    // Go type: time
	    executedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new QueryHistoryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.connID = source["connID"];
	        this.sql = source["sql"];
	        this.status = source["status"];
	        this.errorMessage = source["errorMessage"];
	        this.rowCount = source["rowCount"];
	        this.durationMs = source["durationMs"];
	        this.executedAt = this.convertValues(source["executedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QueryHistoryRequest {
	    connID: number;
	    keyword: string;
	    status: string;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new QueryHistoryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.keyword = source["keyword"];
	        this.status = source["status"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	export class QueryResultDTO {
	    connID: number;
	    sql: string;
	    columns: ColumnDTO[];
	    rows: any[][];
	    rowsAffected: number;
	    lastInsertID: number;
	    durationMs: number;
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QueryResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connID = source["connID"];
	        this.sql = source["sql"];
	        this.columns = this.convertValues(source["columns"], ColumnDTO);
	        this.rows = source["rows"];
	        this.rowsAffected = source["rowsAffected"];
	        this.lastInsertID = source["lastInsertID"];
	        this.durationMs = source["durationMs"];
	        this.truncated = source["truncated"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SchemaColumnDTO {
	    name: string;
	    databaseType: string;
	    nullable: boolean;
	    isPrimary: boolean;
	    autoIncrement: boolean;
	    unsigned: boolean;
	    defaultValue?: string;
	    comment: string;
	    extra: string;
	    ordinalPosition: number;
	    typeLength?: number;
	    precision?: number;
	    scale?: number;
	    collation: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaColumnDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.databaseType = source["databaseType"];
	        this.nullable = source["nullable"];
	        this.isPrimary = source["isPrimary"];
	        this.autoIncrement = source["autoIncrement"];
	        this.unsigned = source["unsigned"];
	        this.defaultValue = source["defaultValue"];
	        this.comment = source["comment"];
	        this.extra = source["extra"];
	        this.ordinalPosition = source["ordinalPosition"];
	        this.typeLength = source["typeLength"];
	        this.precision = source["precision"];
	        this.scale = source["scale"];
	        this.collation = source["collation"];
	    }
	}
	export class SchemaDTO {
	    name: string;
	    owner: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.owner = source["owner"];
	    }
	}
	export class TableDTO {
	    name: string;
	    type: string;
	    comment: string;
	    rowCount: number;
	
	    static createFrom(source: any = {}) {
	        return new TableDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.comment = source["comment"];
	        this.rowCount = source["rowCount"];
	    }
	}
	export class TableDetailDTO {
	    name: string;
	    schema: string;
	    database: string;
	    comment: string;
	    columns: SchemaColumnDTO[];
	    indexes: IndexDTO[];
	    foreignKeys: ForeignKeyDTO[];
	
	    static createFrom(source: any = {}) {
	        return new TableDetailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.schema = source["schema"];
	        this.database = source["database"];
	        this.comment = source["comment"];
	        this.columns = this.convertValues(source["columns"], SchemaColumnDTO);
	        this.indexes = this.convertValues(source["indexes"], IndexDTO);
	        this.foreignKeys = this.convertValues(source["foreignKeys"], ForeignKeyDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TestConnectionRequest {
	    type: string;
	    host: string;
	    port: number;
	    user: string;
	    password: string;
	    database: string;
	    ssl: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TestConnectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.ssl = source["ssl"];
	    }
	}
	export class TestConnectionResultDTO {
	    success: boolean;
	    message: string;
	    serverVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new TestConnectionResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.serverVersion = source["serverVersion"];
	    }
	}
	export class UpdateConnectionRequest {
	    name: string;
	    type: string;
	    host: string;
	    port: number;
	    user: string;
	    password: string;
	    database: string;
	    ssl: boolean;
	    group: string;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateConnectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.ssl = source["ssl"];
	        this.group = source["group"];
	        this.color = source["color"];
	    }
	}
	export class ViewDTO {
	    name: string;
	    definition: string;
	
	    static createFrom(source: any = {}) {
	        return new ViewDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.definition = source["definition"];
	    }
	}

}

