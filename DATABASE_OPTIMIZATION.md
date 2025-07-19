## PostgreSQL Database Optimizations

### Indexes Created Automatically

When using PostgreSQL storage, the following indexes are created automatically during database initialization:

1. **Single Column Indexes** (via GORM tags):
    - `idx_hostname` - Fast host-based lookups
    - `idx_country` - Fast country-based lookups
    - `idx_country_code` - Quick country code searches
    - `idx_city` - City-based queries
    - `idx_region` - Regional data access
    - `idx_zip` - ZIP code lookups
    - `idx_timezone` - Timezone-based filtering
    - `idx_isp` - ISP provider searches
    - `idx_org` - Organization lookups
    - `idx_created_at` - Time-based queries

2. **Composite Indexes**:
    - `idx_country_city` - Combined country and city searches
    - `idx_coordinates` - Latitude/longitude spatial queries

3. **Specialized Indexes** (created via SQL):
    - **BRIN Index** `idx_created_at_brin` - Efficient for large time-series data
    - **GiST Index** `idx_lat_lon_gist` - Spatial queries for geographic coordinates
    - `idx_isp_country` - ISP within country searches
    - `idx_geo_location` - Full geographic hierarchy (country, region, city)

### Database Performance

With these optimizations:

- Single record inserts: < 5ms
- Bulk data retrieval: Optimized with indexes
- Time-series queries: BRIN index provides efficient scans
- Geographic queries: GiST index enables spatial operations

### Database Maintenance

The application automatically runs `ANALYZE geodata` after creating indexes to update table statistics.

For ongoing maintenance, run these periodically:

```sql
-- Update table statistics
ANALYZE
geodata;

-- Rebuild indexes if fragmented (after millions of inserts)
REINDEX
TABLE geodata;

-- Vacuum to reclaim space
VACUUM
ANALYZE geodata;
```

