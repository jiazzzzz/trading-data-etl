const { createApp } = Vue;

const API_BASE = 'http://localhost:8080/api';

createApp({
    data() {
        return {
            loading: false,
            darkMode: false,
            activeTab: 'stocks',
            searchQuery: '',
            selectedDate: '',
            availableDates: [],
            currentPage: 1,
            pageSize: 20,
            exchangeFilters: {
                sh: true,
                sz: true,
                bj: false
            },
            stockSortBy: null,
            stockSortOrder: 'desc',
            
            stats: {
                totalStocks: 0,
                gainers: 0,
                losers: 0,
                tables: 0
            },
            
            tabs: [
                { id: 'stocks', name: '股票列表', icon: 'fas fa-list' },
                { id: 'gainers', name: '涨幅榜', icon: 'fas fa-arrow-up' },
                { id: 'losers', name: '跌幅榜', icon: 'fas fa-arrow-down' },
                { id: 'analytics', name: '数据分析', icon: 'fas fa-chart-pie' },
                { id: 'strategy', name: '策略选股', icon: 'fas fa-magic' }
            ],
            
            topGainers: [],
            topLosers: [],
            stockList: [],
            allStocks: [],
            industryData: {},
            
            // Filter parameters
            filterDays: 7,
            filterMinChange: 7,
            filterMaxChange: 100,
            filterMaxMktcap: 1000,
            filteredStocks: [],
            filterLoading: false,
            filterSortBy: null,
            filterSortOrder: 'desc',
            boardFilters: {
                sh: true,        // 沪市主板
                star: true,      // 科创板
                szMain: true,    // 深主板
                sme: true,       // 中小板
                chinext: true,   // 创业板
                bj: true         // 北证
            },
            
            // Strategy scanner parameters
            strategyDate: '',
            strategyVolumeMultiplier: 2.0,
            strategyMinChangeIncrease: 5.0,
            strategyMinTurnover: 5.0,
            strategyMaxMktcap: 0,
            strategyResults: [],
            strategyLoading: false,
            strategySortBy: 'volume_ratio',
            strategySortOrder: 'desc',
            predefinedStrategies: [
                {
                    id: 'volume_surge',
                    name: '成交量倍增+涨幅加速',
                    description: '当日成交量≥前日2倍 + 当日涨幅-前日涨幅≥5% + 换手率≥5%',
                    params: { volumeMultiplier: 2.0, minChangeIncrease: 5.0, minTurnover: 5.0, maxMktcap: 0 }
                },
                {
                    id: 'small_cap_surge',
                    name: '小盘股放量突破',
                    description: '总市值≤100亿 + 成交量≥前日3倍 + 涨幅加速≥3%',
                    params: { volumeMultiplier: 3.0, minChangeIncrease: 3.0, minTurnover: 3.0, maxMktcap: 100 }
                },
                {
                    id: 'strong_momentum',
                    name: '强势动能股',
                    description: '成交量≥前日1.5倍 + 涨幅加速≥7% + 换手率≥8%',
                    params: { volumeMultiplier: 1.5, minChangeIncrease: 7.0, minTurnover: 8.0, maxMktcap: 0 }
                },
                {
                    id: 'conservative',
                    name: '稳健放量',
                    description: '成交量≥前日1.5倍 + 涨幅加速≥2% + 换手率≥3%',
                    params: { volumeMultiplier: 1.5, minChangeIncrease: 2.0, minTurnover: 3.0, maxMktcap: 0 }
                }
            ],
            
            gainersChartInstance: null,
            losersChartInstance: null,
            industryChartInstance: null
        };
    },
    
    computed: {
        filteredStockList() {
            let filtered = this.allStocks;
            
            // Apply exchange filter first
            filtered = filtered.filter(stock => {
                const tsCode = (stock.ts_code || '').toUpperCase();
                if (tsCode.endsWith('.SH') && this.exchangeFilters.sh) return true;
                if (tsCode.endsWith('.SZ') && this.exchangeFilters.sz) return true;
                if (tsCode.endsWith('.BJ') && this.exchangeFilters.bj) return true;
                return false;
            });
            
            // Apply search filter
            if (this.searchQuery.trim()) {
                const query = this.searchQuery.toLowerCase();
                filtered = filtered.filter(stock => {
                    return (
                        stock.symbol.toLowerCase().includes(query) ||
                        stock.name.toLowerCase().includes(query) ||
                        (stock.pinyin && stock.pinyin.toLowerCase().includes(query))
                    );
                });
            }
            
            // Apply sorting
            if (this.stockSortBy) {
                filtered.sort((a, b) => {
                    let aVal, bVal;
                    
                    switch(this.stockSortBy) {
                        case 'trade':
                            aVal = a.trade || 0;
                            bVal = b.trade || 0;
                            break;
                        case 'changepercent':
                            aVal = a.changepercent || 0;
                            bVal = b.changepercent || 0;
                            break;
                        case 'mktcap':
                            aVal = a.mktcap || 0;
                            bVal = b.mktcap || 0;
                            break;
                        case 'turnoverratio':
                            aVal = a.turnoverratio || 0;
                            bVal = b.turnoverratio || 0;
                            break;
                        default:
                            return 0;
                    }
                    
                    if (this.stockSortOrder === 'asc') {
                        return aVal - bVal;
                    } else {
                        return bVal - aVal;
                    }
                });
            }
            
            // If search query, return all filtered results
            if (this.searchQuery.trim()) {
                return filtered;
            }
            
            // If no search query, apply pagination
            const start = (this.currentPage - 1) * this.pageSize;
            const end = start + this.pageSize;
            return filtered.slice(start, end);
        },
        
        allExchangesSelected() {
            return this.exchangeFilters.sh && this.exchangeFilters.sz && this.exchangeFilters.bj;
        },
        
        displayedFilteredStocks() {
            // First apply board filters
            let filtered = this.filteredStocks.filter(stock => {
                const code = stock.stock_code;
                
                // 沪市主板: 600, 601, 603, 605
                if ((code.startsWith('600') || code.startsWith('601') || code.startsWith('603') || code.startsWith('605')) && this.boardFilters.sh) return true;
                
                // 科创板: 688
                if (code.startsWith('688') && this.boardFilters.star) return true;
                
                // 深主板: 000, 001
                if ((code.startsWith('000') || code.startsWith('001')) && this.boardFilters.szMain) return true;
                
                // 中小板: 002, 003
                if ((code.startsWith('002') || code.startsWith('003')) && this.boardFilters.sme) return true;
                
                // 创业板: 300, 301
                if ((code.startsWith('300') || code.startsWith('301')) && this.boardFilters.chinext) return true;
                
                // 北证: 920, 830, 430
                if ((code.startsWith('920') || code.startsWith('830') || code.startsWith('430')) && this.boardFilters.bj) return true;
                
                return false;
            });
            
            // Then apply sorting if needed
            if (this.filterSortBy && filtered.length > 0) {
                filtered.sort((a, b) => {
                    let aVal, bVal;
                    
                    switch(this.filterSortBy) {
                        case 'max_change':
                            aVal = a.max_change_percent || 0;
                            bVal = b.max_change_percent || 0;
                            break;
                        case 'trigger_count':
                            aVal = a.trigger_dates ? a.trigger_dates.split(',').length : 0;
                            bVal = b.trigger_dates ? b.trigger_dates.split(',').length : 0;
                            break;
                        case 'latest_price':
                            aVal = a.latest_price || 0;
                            bVal = b.latest_price || 0;
                            break;
                        case 'mktcap':
                            aVal = a.mktcap || 0;
                            bVal = b.mktcap || 0;
                            break;
                        default:
                            return 0;
                    }
                    
                    if (this.filterSortOrder === 'asc') {
                        return aVal - bVal;
                    } else {
                        return bVal - aVal;
                    }
                });
            }
            
            return filtered;
        },
        
        sortedStrategyResults() {
            if (!this.strategyResults || this.strategyResults.length === 0) {
                return [];
            }
            
            const results = [...this.strategyResults];
            const sortBy = this.strategySortBy;
            const order = this.strategySortOrder;
            
            results.sort((a, b) => {
                let aVal, bVal;
                
                switch(sortBy) {
                    case 'change_acceleration':
                        aVal = a.change_percent - a.prev_change_percent;
                        bVal = b.change_percent - b.prev_change_percent;
                        break;
                    case 'volume_ratio':
                        aVal = a.volume_ratio || 0;
                        bVal = b.volume_ratio || 0;
                        break;
                    case 'turnover_ratio':
                        aVal = a.turnover_ratio || 0;
                        bVal = b.turnover_ratio || 0;
                        break;
                    case 'close':
                        aVal = a.close || 0;
                        bVal = b.close || 0;
                        break;
                    default:
                        return 0;
                }
                
                if (order === 'asc') {
                    return aVal - bVal;
                } else {
                    return bVal - aVal;
                }
            });
            
            return results;
        }
    },
    
    mounted() {
        this.init();
    },
    
    methods: {
        async init() {
            await this.loadTables();
            await this.loadStats();
            await this.loadStockList();
            await this.loadIndustryData();
            
            if (this.availableDates.length > 0) {
                this.selectedDate = this.availableDates[0];
                await this.loadStatsForDate(this.selectedDate);
                await this.loadTopGainers();
            }
        },
        
        async loadTables() {
            try {
                // Get available dates from history data (last 7 days)
                const datesResponse = await fetch(`${API_BASE}/history/dates/available?limit=7`);
                const datesData = await datesResponse.json();
                this.availableDates = datesData.dates || [];
                
                // Get table count
                const tablesResponse = await fetch(`${API_BASE}/tables`);
                const tablesData = await tablesResponse.json();
                this.stats.tables = tablesData.count;
                    
            } catch (error) {
                console.error('Error loading tables:', error);
                this.showError('加载数据表失败');
            }
        },
        
        async loadStats() {
            try {
                const response = await fetch(`${API_BASE}/stats`);
                const data = await response.json();
                this.stats.totalStocks = data.total_stocks || 0;
                this.stats.gainers = data.gainers || 0;
                this.stats.losers = data.losers || 0;
                this.stats.tables = data.tables || 0;
            } catch (error) {
                console.error('Error loading stats:', error);
            }
        },
        
        async loadStatsForDate(date) {
            if (!date) return;
            try {
                const response = await fetch(`${API_BASE}/stats?date=${date}`);
                const data = await response.json();
                this.stats.gainers = data.gainers || 0;
                this.stats.losers = data.losers || 0;
            } catch (error) {
                console.error('Error loading stats for date:', error);
            }
        },
        
        async loadTopLosers() {
            if (!this.selectedDate) {
                console.log('No date selected');
                return;
            }
            
            this.loading = true;
            console.log('Loading top losers for date:', this.selectedDate);
            
            try {
                const url = `${API_BASE}/history/losers/${this.selectedDate}?limit=20`;
                console.log('Fetching:', url);
                
                const response = await fetch(url);
                const data = await response.json();
                
                console.log('Response data:', data);
                
                if (data.error) {
                    console.error('API error:', data.error);
                    this.showError(data.error);
                    this.loading = false;
                    return;
                }
                
                this.topLosers = data.top_losers || [];
                console.log('Top losers loaded:', this.topLosers.length);
                
                if (this.topLosers.length === 0) {
                    console.warn('No top losers data');
                    this.showError('该日期没有数据');
                    this.loading = false;
                    return;
                }
                
                this.loading = false;
                
                // Render chart after loading is done
                this.$nextTick(() => {
                    setTimeout(() => {
                        this.renderLosersChart();
                    }, 150);
                });
            } catch (error) {
                console.error('Error loading top losers:', error);
                this.showError('加载跌幅榜失败: ' + error.message);
                this.loading = false;
            }
        },
        
        async loadTopGainers() {
            if (!this.selectedDate) {
                console.log('No date selected');
                return;
            }
            
            this.loading = true;
            console.log('Loading top gainers for date:', this.selectedDate);
            
            try {
                const url = `${API_BASE}/history/gainers/${this.selectedDate}?limit=20`;
                console.log('Fetching:', url);
                
                const response = await fetch(url);
                const data = await response.json();
                
                console.log('Response data:', data);
                
                if (data.error) {
                    console.error('API error:', data.error);
                    this.showError(data.error);
                    this.loading = false;
                    return;
                }
                
                this.topGainers = data.top_gainers || [];
                console.log('Top gainers loaded:', this.topGainers.length);
                
                if (this.topGainers.length === 0) {
                    console.warn('No top gainers data');
                    this.showError('该日期没有数据');
                    this.loading = false;
                    return;
                }
                
                this.loading = false;
                
                // Render chart after loading is done
                this.$nextTick(() => {
                    setTimeout(() => {
                        this.renderGainersChart();
                    }, 150);
                });
            } catch (error) {
                console.error('Error loading top gainers:', error);
                this.showError('加载涨幅榜失败: ' + error.message);
                this.loading = false;
            }
        },
        
        async loadStockList() {
            this.loading = true;
            try {
                // Load all stocks for filtering (only once)
                if (this.allStocks.length === 0) {
                    const response = await fetch(`${API_BASE}/stocks?limit=10000`);
                    const data = await response.json();
                    this.allStocks = data.stocks || [];
                    console.log('Loaded stocks:', this.allStocks.length);
                    console.log('Sample stock:', this.allStocks[0]);
                }
            } catch (error) {
                console.error('Error loading stock list:', error);
                this.showError('加载股票列表失败');
            } finally {
                this.loading = false;
            }
        },
        
        filterStocks() {
            // Filter is handled by computed property
        },
        
        clearSearch() {
            this.searchQuery = '';
        },
        
        sortStocks(column) {
            if (this.stockSortBy === column) {
                // Toggle sort order
                this.stockSortOrder = this.stockSortOrder === 'desc' ? 'asc' : 'desc';
            } else {
                // New column, default to descending
                this.stockSortBy = column;
                this.stockSortOrder = 'desc';
            }
        },
        
        async applyFilter() {
            this.filterLoading = true;
            try {
                const url = `${API_BASE}/filter/stocks?days=${this.filterDays}&min_change=${this.filterMinChange}&max_change=${this.filterMaxChange}&max_mktcap=${this.filterMaxMktcap}&limit=500`;
                console.log('Fetching filter:', url);
                const response = await fetch(url);
                const data = await response.json();
                
                console.log('Filter response:', data);
                
                if (data.error) {
                    this.showError(data.error);
                    this.filteredStocks = [];
                } else {
                    this.filteredStocks = data.filtered_stocks || [];
                    if (this.filteredStocks.length === 0) {
                        this.showError('未找到符合条件的股票，请调整筛选条件');
                    }
                }
            } catch (error) {
                console.error('Error filtering stocks:', error);
                this.showError('筛选失败: ' + error.message);
                this.filteredStocks = [];
            } finally {
                this.filterLoading = false;
            }
        },
        

        sortFilteredStocks(column) {
            if (this.filterSortBy === column) {
                // Toggle sort order
                this.filterSortOrder = this.filterSortOrder === 'desc' ? 'asc' : 'desc';
            } else {
                // New column, default to descending
                this.filterSortBy = column;
                this.filterSortOrder = 'desc';
            }
        },
        
        clearFilter() {
            this.filterDays = 7;
            this.filterMinChange = 7;
            this.filterMaxChange = 100;
            this.filterMaxMktcap = 1000;
            this.filteredStocks = [];
        },
        
        applyPredefinedStrategy(strategy) {
            this.strategyVolumeMultiplier = strategy.params.volumeMultiplier;
            this.strategyMinChangeIncrease = strategy.params.minChangeIncrease;
            this.strategyMinTurnover = strategy.params.minTurnover;
            this.strategyMaxMktcap = strategy.params.maxMktcap;
            this.runStrategyScanner();
        },
        
        async runStrategyScanner() {
            this.strategyLoading = true;
            try {
                const params = new URLSearchParams({
                    date: this.strategyDate || '',
                    volume_multiplier: this.strategyVolumeMultiplier,
                    min_change_increase: this.strategyMinChangeIncrease,
                    min_turnover: this.strategyMinTurnover,
                    max_mktcap: this.strategyMaxMktcap,
                    limit: 100
                });
                
                const url = `${API_BASE}/strategy/scan?${params}`;
                console.log('Running strategy scan:', url);
                
                const response = await fetch(url);
                const data = await response.json();
                
                console.log('Strategy results:', data);
                
                if (data.error) {
                    this.showError(data.error);
                    this.strategyResults = [];
                } else {
                    this.strategyResults = data.results || [];
                    this.strategyDate = data.date;
                    if (this.strategyResults.length === 0) {
                        this.showError('未找到符合条件的股票，请调整策略参数');
                    }
                }
            } catch (error) {
                console.error('Error running strategy:', error);
                this.showError('策略扫描失败: ' + error.message);
                this.strategyResults = [];
            } finally {
                this.strategyLoading = false;
            }
        },
        
        clearStrategy() {
            this.strategyVolumeMultiplier = 2.0;
            this.strategyMinChangeIncrease = 5.0;
            this.strategyMinTurnover = 5.0;
            this.strategyMaxMktcap = 0;
            this.strategyResults = [];
        },
        
        sortStrategyResults(column) {
            if (this.strategySortBy === column) {
                // Toggle sort order if clicking the same column
                this.strategySortOrder = this.strategySortOrder === 'desc' ? 'asc' : 'desc';
            } else {
                // Set new column and default to descending
                this.strategySortBy = column;
                this.strategySortOrder = 'desc';
            }
        },
        
        async loadIndustryData() {
            try {
                const sql = "SELECT industry, COUNT(*) as count FROM stock_list GROUP BY industry ORDER BY count DESC LIMIT 15";
                const response = await fetch(`${API_BASE}/query?sql=${encodeURIComponent(sql)}`);
                const data = await response.json();
                
                if (data.results) {
                    this.industryData = data.results.reduce((acc, item) => {
                        acc[item.industry] = item.count;
                        return acc;
                    }, {});
                    
                    this.renderIndustryChart();
                }
            } catch (error) {
                console.error('Error loading industry data:', error);
            }
        },
        
        renderGainersChart() {
            console.log('=== renderGainersChart called ===');
            this.$nextTick(() => {
                // Check if echarts is loaded
                if (typeof echarts === 'undefined') {
                    console.error('ECharts library not loaded yet');
                    return;
                }
                
                const chartDom = this.$refs.gainersChart;
                if (!chartDom) {
                    console.error('Chart DOM not found');
                    return;
                }
                
                if (this.topGainers.length === 0) {
                    console.warn('No data to render');
                    return;
                }
                
                console.log('Rendering gainers chart with', this.topGainers.length, 'items');
                console.log('Selected date:', this.selectedDate);
                
                // Dispose old chart instance and create new one
                if (this.gainersChartInstance) {
                    this.gainersChartInstance.dispose();
                }
                this.gainersChartInstance = echarts.init(chartDom);
                
                const names = this.topGainers.map(s => s.stock_name || 'Unknown');
                const changes = this.topGainers.map(s => s.change_percent || 0);
                const prices = this.topGainers.map(s => s.close || 0);
                
                console.log('Chart data:', { names, changes, prices });
                console.log('Sample stock data:', this.topGainers[0]);
                
                // Validate data
                if (names.length === 0 || changes.length === 0) {
                    console.error('Invalid chart data');
                    return;
                }
                
                const option = {
                    title: {
                        text: `涨幅榜 - ${this.formatDate(this.selectedDate)}`,
                        left: 'center',
                        textStyle: {
                            fontSize: 18,
                            fontWeight: 'bold'
                        }
                    },
                    tooltip: {
                        trigger: 'axis',
                        axisPointer: {
                            type: 'shadow'
                        },
                        formatter: (params) => {
                            const index = params[0].dataIndex;
                            const stock = this.topGainers[index];
                            return `
                                <div style="padding: 10px;">
                                    <div style="font-weight: bold; margin-bottom: 5px;">${stock.stock_name} (${stock.stock_code})</div>
                                    <div>涨幅: <span style="color: ${stock.change_percent >= 0 ? '#10b981' : '#ef4444'}; font-weight: bold;">${stock.change_percent.toFixed(2)}%</span></div>
                                    <div>收盘: ¥${stock.close.toFixed(2)}</div>
                                    <div>涨跌: ¥${stock.change.toFixed(2)}</div>
                                    <div>成交量: ${(stock.volume / 10000).toFixed(2)}万手</div>
                                </div>
                            `;
                        }
                    },
                    grid: {
                        left: '3%',
                        right: '4%',
                        bottom: '3%',
                        containLabel: true
                    },
                    xAxis: {
                        type: 'value',
                        name: '涨幅 (%)',
                        axisLabel: {
                            formatter: '{value}%'
                        }
                    },
                    yAxis: {
                        type: 'category',
                        data: names,
                        axisLabel: {
                            interval: 0,
                            fontSize: 11
                        }
                    },
                    series: [
                        {
                            name: '涨幅',
                            type: 'bar',
                            data: changes,
                            itemStyle: {
                                color: (params) => {
                                    return params.value >= 0 ? '#10b981' : '#ef4444';
                                },
                                borderRadius: [0, 4, 4, 0]
                            },
                            label: {
                                show: true,
                                position: 'right',
                                formatter: function(params) {
                                    return parseFloat(params.value).toFixed(2) + '%';
                                },
                                fontSize: 10
                            }
                        }
                    ]
                };
                
                try {
                    this.gainersChartInstance.setOption(option);
                    console.log('Chart rendered successfully');
                } catch (error) {
                    console.error('Error rendering chart:', error);
                    console.error('Chart option:', option);
                }
            });
        },
        
        renderLosersChart() {
            this.$nextTick(() => {
                // Check if echarts is loaded
                if (typeof echarts === 'undefined') {
                    console.error('ECharts library not loaded yet');
                    return;
                }
                
                const chartDom = this.$refs.losersChart;
                if (!chartDom) {
                    console.error('Chart DOM not found');
                    return;
                }
                
                if (this.topLosers.length === 0) {
                    console.warn('No data to render');
                    return;
                }
                
                console.log('Rendering losers chart with', this.topLosers.length, 'items');
                
                // Dispose old chart instance and create new one
                if (this.losersChartInstance) {
                    this.losersChartInstance.dispose();
                }
                this.losersChartInstance = echarts.init(chartDom);
                
                const names = this.topLosers.map(s => s.stock_name || 'Unknown');
                const changes = this.topLosers.map(s => s.change_percent || 0);
                
                console.log('Chart data:', { names, changes });
                console.log('Sample stock data:', this.topLosers[0]);
                
                // Validate data
                if (names.length === 0 || changes.length === 0) {
                    console.error('Invalid chart data');
                    return;
                }
                
                const option = {
                    title: {
                        text: `跌幅榜 - ${this.formatDate(this.selectedDate)}`,
                        left: 'center',
                        textStyle: {
                            fontSize: 18,
                            fontWeight: 'bold'
                        }
                    },
                    tooltip: {
                        trigger: 'axis',
                        axisPointer: {
                            type: 'shadow'
                        },
                        formatter: (params) => {
                            const index = params[0].dataIndex;
                            const stock = this.topLosers[index];
                            return `
                                <div style="padding: 10px;">
                                    <div style="font-weight: bold; margin-bottom: 5px;">${stock.stock_name} (${stock.stock_code})</div>
                                    <div>跌幅: <span style="color: ${stock.change_percent >= 0 ? '#10b981' : '#ef4444'}; font-weight: bold;">${stock.change_percent.toFixed(2)}%</span></div>
                                    <div>收盘: ¥${stock.close.toFixed(2)}</div>
                                    <div>涨跌: ¥${stock.change.toFixed(2)}</div>
                                    <div>成交量: ${(stock.volume / 10000).toFixed(2)}万手</div>
                                </div>
                            `;
                        }
                    },
                    grid: {
                        left: '3%',
                        right: '4%',
                        bottom: '3%',
                        containLabel: true
                    },
                    xAxis: {
                        type: 'value',
                        name: '跌幅 (%)',
                        axisLabel: {
                            formatter: '{value}%'
                        }
                    },
                    yAxis: {
                        type: 'category',
                        data: names,
                        axisLabel: {
                            interval: 0,
                            fontSize: 11
                        }
                    },
                    series: [
                        {
                            name: '跌幅',
                            type: 'bar',
                            data: changes,
                            itemStyle: {
                                color: (params) => {
                                    return params.value >= 0 ? '#10b981' : '#ef4444';
                                },
                                borderRadius: [0, 4, 4, 0]
                            },
                            label: {
                                show: true,
                                position: 'right',
                                formatter: function(params) {
                                    return parseFloat(params.value).toFixed(2) + '%';
                                },
                                fontSize: 10
                            }
                        }
                    ]
                };
                
                try {
                    this.losersChartInstance.setOption(option);
                    console.log('Losers chart rendered successfully');
                } catch (error) {
                    console.error('Error rendering losers chart:', error);
                    console.error('Chart option:', option);
                }
            });
        },
        
        renderIndustryChart() {
            this.$nextTick(() => {
                // Check if echarts is loaded
                if (typeof echarts === 'undefined') {
                    console.error('ECharts library not loaded yet');
                    return;
                }
                
                const chartDom = this.$refs.industryChart;
                if (!chartDom) return;
                
                if (!this.industryChartInstance) {
                    this.industryChartInstance = echarts.init(chartDom);
                }
                
                const data = Object.entries(this.industryData).map(([name, value]) => ({
                    name,
                    value
                }));
                
                const option = {
                    title: {
                        text: '行业分布',
                        subtext: '按股票数量统计',
                        left: 'center',
                        textStyle: {
                            fontSize: 18,
                            fontWeight: 'bold'
                        }
                    },
                    tooltip: {
                        trigger: 'item',
                        formatter: '{b}: {c} ({d}%)'
                    },
                    legend: {
                        orient: 'vertical',
                        left: 'left',
                        top: 'middle',
                        textStyle: {
                            fontSize: 12
                        }
                    },
                    series: [
                        {
                            name: '行业',
                            type: 'pie',
                            radius: ['40%', '70%'],
                            center: ['60%', '50%'],
                            avoidLabelOverlap: false,
                            itemStyle: {
                                borderRadius: 10,
                                borderColor: '#fff',
                                borderWidth: 2
                            },
                            label: {
                                show: true,
                                formatter: '{b}\n{d}%'
                            },
                            emphasis: {
                                label: {
                                    show: true,
                                    fontSize: 14,
                                    fontWeight: 'bold'
                                }
                            },
                            data: data
                        }
                    ]
                };
                
                this.industryChartInstance.setOption(option);
            });
        },
        
        async refreshData() {
            await this.init();
        },
        
        toggleTheme() {
            this.darkMode = !this.darkMode;
            // Implement dark mode if needed
        },
        
        nextPage() {
            this.currentPage++;
        },
        
        prevPage() {
            if (this.currentPage > 1) {
                this.currentPage--;
            }
        },
        
        formatDate(dateStr) {
            if (!dateStr) return '';
            // Format YYYYMMDD to YYYY-MM-DD
            return `${dateStr.slice(0, 4)}-${dateStr.slice(4, 6)}-${dateStr.slice(6, 8)}`;
        },
        
        formatListDate(dateStr) {
            if (!dateStr) return '';
            return this.formatDate(dateStr);
        },
        
        formatTriggerDates(dates) {
            if (!dates || dates.length === 0) return '-';
            return dates.map(d => this.formatDate(d)).join('; ');
        },
        
        getXueqiuSymbol(stockCode) {
            // Convert stock code to Xueqiu format (SH/SZ prefix)
            if (!stockCode) return '';
            if (stockCode.startsWith('6')) {
                return 'SH' + stockCode;
            } else if (stockCode.startsWith('0') || stockCode.startsWith('3')) {
                return 'SZ' + stockCode;
            } else if (stockCode.startsWith('8') || stockCode.startsWith('4')) {
                return 'BJ' + stockCode;
            }
            return 'SZ' + stockCode; // Default to SZ
        },
        
        showError(message) {
            alert(message);
        }
    },
    
    watch: {
        activeTab(newTab) {
            if (newTab === 'gainers') {
                // Re-render gainers chart when switching to tab
                this.$nextTick(() => {
                    if (this.topGainers.length > 0) {
                        this.renderGainersChart();
                    } else {
                        this.loadTopGainers();
                    }
                });
            } else if (newTab === 'losers') {
                // Re-render losers chart when switching to tab
                this.$nextTick(() => {
                    if (this.topLosers.length > 0) {
                        this.renderLosersChart();
                    } else {
                        this.loadTopLosers();
                    }
                });
            } else if (newTab === 'stocks' && this.stockList.length === 0) {
                this.loadStockList();
            } else if (newTab === 'analytics') {
                this.$nextTick(() => {
                    if (Object.keys(this.industryData).length > 0) {
                        this.renderIndustryChart();
                    } else {
                        this.loadIndustryData();
                    }
                });
            }
        },
        
        selectedDate(newDate) {
            // Reload data when date changes
            if (this.activeTab === 'gainers') {
                this.loadTopGainers();
            } else if (this.activeTab === 'losers') {
                this.loadTopLosers();
            }
            // Update stats for the selected date
            this.loadStatsForDate(newDate);
        }
    }
}).mount('#app');
