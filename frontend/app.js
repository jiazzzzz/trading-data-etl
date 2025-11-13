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
            boardFiltersStock: {
                shMain: true,    // 沪市主板
                star: true,      // 科创板
                szMain: true,    // 深主板
                sme: true,       // 中小板
                chinext: true,   // 创业板
                bj: false        // 北证
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
                { id: 'watchlist', name: '我的自选股', icon: 'fas fa-star' },
                { id: 'warninglist', name: '问题/预警股', icon: 'fas fa-exclamation-triangle' },
                { id: 'gainers', name: '涨幅榜', icon: 'fas fa-arrow-up' },
                { id: 'losers', name: '跌幅榜', icon: 'fas fa-arrow-down' },
                { id: 'analytics', name: '数据分析', icon: 'fas fa-chart-pie' },
                { id: 'strategy', name: '策略选股', icon: 'fas fa-magic' }
            ],
            
            topGainers: [],
            topLosers: [],
            stockList: [],
            allStocks: [],
            watchlist: [],
            watchlistCodes: [],
            warninglist: [],
            warninglistCodes: [],
            allTags: [],
            selectedTagFilters: [], // Changed to array for multi-select
            industryData: {},
            
            // Tag management for stock
            showTagModal: false,
            currentStockForTag: null,
            stockTags: [],
            newTagName: '',
            newTagColor: '#3B82F6',
            
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
            
            // Strategy cache for watchlist matching
            strategyCacheDate: null,
            strategyCachedResults: {},
            
            // Sparkline cache (use object for Vue reactivity)
            loadedSparklines: {},
            
            // Toast notifications
            toasts: [],
            toastIdCounter: 0,
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
            // If filtering by tag, use allStocks
            if (this.selectedTagFilters.length > 0) {
                return this.allStocks;
            }
            
            // Determine which data source to use
            let source = this.searchQuery.trim() ? this.allStocks : this.stockList;
            
            if (!source || source.length === 0) {
                return [];
            }
            
            // Backend already filtered by board, just apply search filter if needed
            let filtered = source;
            
            if (this.searchQuery.trim()) {
                const query = this.searchQuery.toLowerCase();
                filtered = source.filter(stock => {
                    return (
                        stock.symbol.toLowerCase().includes(query) ||
                        stock.name.toLowerCase().includes(query) ||
                        (stock.pinyin && stock.pinyin.toLowerCase().includes(query))
                    );
                });
            }
            
            // Apply search filter if searching
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
            
            return filtered;
        },
        
        allBoardsSelected() {
            return this.boardFiltersStock.shMain && this.boardFiltersStock.star && 
                   this.boardFiltersStock.szMain && this.boardFiltersStock.sme && 
                   this.boardFiltersStock.chinext && this.boardFiltersStock.bj;
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
        this.loadWatchlist();
        this.init();
    },
    
    methods: {
        async init() {
            await this.loadTables();
            await this.loadStats();
            await this.loadTags();
            await this.loadStockList();
            await this.loadIndustryData();
            
            if (this.availableDates.length > 0) {
                this.selectedDate = this.availableDates[0];
                await this.loadStatsForDate(this.selectedDate);
                await this.loadTopGainers();
            }
        },
        
        async loadTags() {
            try {
                const response = await fetch(`${API_BASE}/tags`);
                const data = await response.json();
                this.allTags = data.tags || [];
            } catch (error) {
                console.error('Error loading tags:', error);
            }
        },
        
        async loadWatchlist() {
            // Load watchlist from database
            try {
                const response = await fetch(`${API_BASE}/watchlist`);
                const data = await response.json();
                this.watchlistCodes = data.codes || [];
            } catch (error) {
                console.error('Error loading watchlist:', error);
                this.watchlistCodes = [];
            }
            
            // Load warninglist from database
            try {
                const response = await fetch(`${API_BASE}/warninglist`);
                const data = await response.json();
                this.warninglistCodes = data.codes || [];
            } catch (error) {
                console.error('Error loading warninglist:', error);
                this.warninglistCodes = [];
            }
        },
        
        async addToWatchlist(stock) {
            const code = stock.symbol || stock.stock_code;
            if (this.watchlistCodes.includes(code)) {
                this.showError('该股票已在自选股中');
                return;
            }
            
            try {
                const response = await fetch(`${API_BASE}/watchlist/${code}`, {
                    method: 'POST'
                });
                const data = await response.json();
                
                if (response.ok) {
                    this.watchlistCodes.push(code);
                    this.showSuccess(`已添加 ${stock.name || stock.stock_name} 到自选股`);
                } else {
                    this.showError(data.error || '添加失败');
                }
            } catch (error) {
                console.error('Error adding to watchlist:', error);
                this.showError('添加失败');
            }
        },
        
        async removeFromWatchlist(code) {
            try {
                const response = await fetch(`${API_BASE}/watchlist/${code}`, {
                    method: 'DELETE'
                });
                const data = await response.json();
                
                if (response.ok) {
                    const index = this.watchlistCodes.indexOf(code);
                    if (index > -1) {
                        this.watchlistCodes.splice(index, 1);
                    }
                    this.loadWatchlistStocks();
                    this.showSuccess('已从自选股移除');
                } else {
                    this.showError(data.error || '移除失败');
                }
            } catch (error) {
                console.error('Error removing from watchlist:', error);
                this.showError('移除失败');
            }
        },
        
        isInWatchlist(stock) {
            const code = stock.symbol || stock.stock_code;
            return this.watchlistCodes.includes(code);
        },
        
        async addToWarninglist(stock) {
            const code = stock.symbol || stock.stock_code;
            if (this.warninglistCodes.includes(code)) {
                this.showError('该股票已在预警股中');
                return;
            }
            
            try {
                const response = await fetch(`${API_BASE}/warninglist/${code}`, {
                    method: 'POST'
                });
                const data = await response.json();
                
                if (response.ok) {
                    this.warninglistCodes.push(code);
                    this.showSuccess(`已添加 ${stock.name || stock.stock_name} 到预警股`);
                } else {
                    this.showError(data.error || '添加失败');
                }
            } catch (error) {
                console.error('Error adding to warninglist:', error);
                this.showError('添加失败');
            }
        },
        
        async removeFromWarninglist(code) {
            try {
                const response = await fetch(`${API_BASE}/warninglist/${code}`, {
                    method: 'DELETE'
                });
                const data = await response.json();
                
                if (response.ok) {
                    const index = this.warninglistCodes.indexOf(code);
                    if (index > -1) {
                        this.warninglistCodes.splice(index, 1);
                    }
                    this.loadWarninglistStocks();
                    this.showSuccess('已从预警股移除');
                } else {
                    this.showError(data.error || '移除失败');
                }
            } catch (error) {
                console.error('Error removing from warninglist:', error);
                this.showError('移除失败');
            }
        },
        
        isInWarninglist(stock) {
            const code = stock.symbol || stock.stock_code;
            return this.warninglistCodes.includes(code);
        },
        
        async refreshStrategyCache() {
            // Force refresh strategy cache and check matches
            this.strategyCacheDate = null;
            this.strategyCachedResults = {};
            
            if (this.activeTab === 'watchlist' && this.watchlist.length > 0) {
                this.loading = true;
                try {
                    await this.checkStrategyMatches();
                    this.showSuccess('策略匹配已更新');
                } catch (error) {
                    console.error('Error refreshing strategy cache:', error);
                    this.showError('刷新策略失败');
                } finally {
                    this.loading = false;
                }
            }
        },
        
        async loadWatchlistStocks() {
            if (this.watchlistCodes.length === 0) {
                this.watchlist = [];
                return;
            }
            
            this.loading = true;
            try {
                // Build SQL to fetch watchlist stocks
                const codes = this.watchlistCodes.map(c => `'${c}'`).join(',');
                const sql = `SELECT ts_code, symbol, name, pinyin FROM stock_list WHERE symbol IN (${codes})`;
                const response = await fetch(`${API_BASE}/query?sql=${encodeURIComponent(sql)}`);
                const data = await response.json();
                
                if (data.results) {
                    // Preserve existing matchedStrategies if they exist
                    const existingMatches = {};
                    if (this.watchlist && this.watchlist.length > 0) {
                        this.watchlist.forEach(stock => {
                            if (stock.matchedStrategies && stock.matchedStrategies.length > 0) {
                                existingMatches[stock.symbol] = stock.matchedStrategies;
                            }
                        });
                    }
                    
                    this.watchlist = data.results;
                    
                    // Restore matchedStrategies or initialize as empty
                    this.watchlist.forEach(stock => {
                        stock.matchedStrategies = existingMatches[stock.symbol] || [];
                    });
                    
                    // Try to get latest trading data
                    const latestTableResponse = await fetch(`${API_BASE}/tables`);
                    const tablesData = await latestTableResponse.json();
                    
                    if (tablesData.tables && tablesData.tables.length > 0) {
                        const dailyTables = tablesData.tables.filter(t => t.name.startsWith('stock_daily_'));
                        if (dailyTables.length > 0) {
                            const latestTable = dailyTables[dailyTables.length - 1].name;
                            
                            // Fetch trading data for watchlist stocks
                            for (let stock of this.watchlist) {
                                try {
                                    const tradeSql = `SELECT trade, changepercent, mktcap, turnoverratio FROM ${latestTable} WHERE REPLACE(REPLACE(REPLACE(symbol, 'sh', ''), 'sz', ''), 'bj', '') = '${stock.symbol}' LIMIT 1`;
                                    const tradeResponse = await fetch(`${API_BASE}/query?sql=${encodeURIComponent(tradeSql)}`);
                                    const tradeData = await tradeResponse.json();
                                    
                                    if (tradeData.results && tradeData.results.length > 0) {
                                        Object.assign(stock, tradeData.results[0]);
                                    }
                                } catch (e) {
                                    console.error('Error loading trade data for', stock.symbol, e);
                                }
                            }
                            
                            // Don't check strategy matches automatically
                            // User needs to click "刷新策略" button to check
                        }
                    }
                }
            } catch (error) {
                console.error('Error loading watchlist:', error);
                this.showError('加载自选股失败');
            } finally {
                this.loading = false;
            }
        },
        
        async checkStrategyMatches() {
            // Get current date for cache key
            const today = new Date().toISOString().split('T')[0].replace(/-/g, '');
            
            // Check if cache is valid (same day)
            if (this.strategyCacheDate !== today) {
                console.log('Fetching fresh strategy data...');
                this.strategyCachedResults = {};
                this.strategyCacheDate = today;
                
                // Fetch all strategies once and cache results
                for (let strategy of this.predefinedStrategies) {
                    try {
                        const params = strategy.params;
                        const url = `${API_BASE}/strategy/scan?date=&volume_multiplier=${params.volumeMultiplier}&min_change_increase=${params.minChangeIncrease}&min_turnover=${params.minTurnover}&max_mktcap=${params.maxMktcap}&limit=1000`;
                        
                        const response = await fetch(url);
                        const data = await response.json();
                        
                        if (data.results) {
                            // Cache results by stock code for fast lookup
                            this.strategyCachedResults[strategy.id] = {};
                            for (let result of data.results) {
                                this.strategyCachedResults[strategy.id][result.stock_code] = result;
                            }
                        }
                    } catch (e) {
                        console.error('Error fetching strategy', strategy.name, e);
                    }
                }
                console.log('Strategy cache updated');
            } else {
                console.log('Using cached strategy data');
            }
            
            // Now check each watchlist stock against cached results (fast)
            for (let i = 0; i < this.watchlist.length; i++) {
                const stock = this.watchlist[i];
                const matches = [];
                
                for (let strategy of this.predefinedStrategies) {
                    const cachedResults = this.strategyCachedResults[strategy.id];
                    if (cachedResults && cachedResults[stock.symbol]) {
                        matches.push({
                            name: strategy.name,
                            description: strategy.description,
                            data: cachedResults[stock.symbol]
                        });
                    }
                }
                
                // Use Vue.set or direct assignment to trigger reactivity
                this.watchlist[i].matchedStrategies = matches;
            }
            
            // Force Vue to update the view
            this.$forceUpdate();
        },
        
        async loadWarninglistStocks() {
            if (this.warninglistCodes.length === 0) {
                this.warninglist = [];
                return;
            }
            
            this.loading = true;
            try {
                // Build SQL to fetch warninglist stocks
                const codes = this.warninglistCodes.map(c => `'${c}'`).join(',');
                const sql = `SELECT ts_code, symbol, name, pinyin FROM stock_list WHERE symbol IN (${codes})`;
                const response = await fetch(`${API_BASE}/query?sql=${encodeURIComponent(sql)}`);
                const data = await response.json();
                
                if (data.results) {
                    this.warninglist = data.results;
                    
                    // Try to get latest trading data
                    const latestTableResponse = await fetch(`${API_BASE}/tables`);
                    const tablesData = await latestTableResponse.json();
                    
                    if (tablesData.tables && tablesData.tables.length > 0) {
                        const dailyTables = tablesData.tables.filter(t => t.name.startsWith('stock_daily_'));
                        if (dailyTables.length > 0) {
                            const latestTable = dailyTables[dailyTables.length - 1].name;
                            
                            // Fetch trading data for warninglist stocks
                            for (let stock of this.warninglist) {
                                try {
                                    const tradeSql = `SELECT trade, changepercent, mktcap, turnoverratio FROM ${latestTable} WHERE REPLACE(REPLACE(REPLACE(symbol, 'sh', ''), 'sz', ''), 'bj', '') = '${stock.symbol}' LIMIT 1`;
                                    const tradeResponse = await fetch(`${API_BASE}/query?sql=${encodeURIComponent(tradeSql)}`);
                                    const tradeData = await tradeResponse.json();
                                    
                                    if (tradeData.results && tradeData.results.length > 0) {
                                        Object.assign(stock, tradeData.results[0]);
                                    }
                                } catch (e) {
                                    console.error('Error loading trade data for', stock.symbol, e);
                                }
                            }
                        }
                    }
                }
            } catch (error) {
                console.error('Error loading warninglist:', error);
                this.showError('加载预警股失败');
            } finally {
                this.loading = false;
            }
        },
        
        showSuccess(message) {
            this.showToast(message, 'success');
        },
        
        showToast(message, type = 'info') {
            const id = this.toastIdCounter++;
            const toast = { id, message, type };
            this.toasts.push(toast);
            
            // Auto remove after 3 seconds
            setTimeout(() => {
                this.removeToast(id);
            }, 3000);
        },
        
        removeToast(id) {
            const index = this.toasts.findIndex(t => t.id === id);
            if (index > -1) {
                this.toasts.splice(index, 1);
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
                // If filtering by multiple tags, load stocks that have ANY of the selected tags
                if (this.selectedTagFilters.length > 0) {
                    // Fetch stocks for each tag and merge results (OR logic)
                    const stocksMap = new Map();
                    for (const tagId of this.selectedTagFilters) {
                        const response = await fetch(`${API_BASE}/tags/${tagId}/stocks?limit=10000`);
                        if (!response.ok) continue;
                        const data = await response.json();
                        const stocks = data.stocks || [];
                        stocks.forEach(stock => {
                            stocksMap.set(stock.symbol, stock);
                        });
                    }
                    this.allStocks = Array.from(stocksMap.values());
                    this.stockList = [];
                    this.loading = false;
                    return;
                }
                
                // Build board filter parameter (always send it)
                const boards = [];
                if (this.boardFiltersStock.shMain) boards.push('shMain');
                if (this.boardFiltersStock.star) boards.push('star');
                if (this.boardFiltersStock.szMain) boards.push('szMain');
                if (this.boardFiltersStock.sme) boards.push('sme');
                if (this.boardFiltersStock.chinext) boards.push('chinext');
                if (this.boardFiltersStock.bj) boards.push('bj');
                const boardParam = `&boards=${boards.join(',')}`;
                
                if (this.searchQuery.trim()) {
                    // Load all stocks for client-side search filtering
                    if (this.allStocks.length === 0) {
                        const response = await fetch(`${API_BASE}/stocks?limit=10000${boardParam}`);
                        const data = await response.json();
                        this.allStocks = data.stocks || [];
                    }
                } else {
                    // Load current page from server with board filter
                    const offset = (this.currentPage - 1) * this.pageSize;
                    const response = await fetch(`${API_BASE}/stocks?limit=${this.pageSize}&offset=${offset}${boardParam}`);
                    const data = await response.json();
                    this.stockList = data.stocks || [];
                }
            } catch (error) {
                console.error('Error loading stock list:', error);
                this.showError('加载股票列表失败');
            } finally {
                this.loading = false;
            }
        },
        
        filterByTag(tagId) {
            // Toggle tag selection
            const index = this.selectedTagFilters.indexOf(tagId);
            if (index > -1) {
                this.selectedTagFilters.splice(index, 1);
            } else {
                this.selectedTagFilters.push(tagId);
            }
            this.searchQuery = '';
            this.currentPage = 1;
            this.loadStockList();
        },
        
        clearTagFilter() {
            this.selectedTagFilters = [];
            this.allStocks = [];
            this.currentPage = 1;
            this.loadStockList();
        },
        
        // Tag management for individual stocks
        async openTagModal(stock) {
            this.currentStockForTag = stock;
            this.showTagModal = true;
            this.newTagName = '';
            await this.loadStockTags(stock.symbol);
        },
        
        closeTagModal() {
            this.showTagModal = false;
            this.currentStockForTag = null;
            this.stockTags = [];
            this.newTagName = '';
        },
        
        async loadStockTags(stockCode) {
            try {
                const response = await fetch(`${API_BASE}/stock-tags/${stockCode}`);
                const data = await response.json();
                this.stockTags = data.tags || [];
            } catch (error) {
                console.error('Failed to load stock tags:', error);
                this.stockTags = [];
            }
        },
        
        async addTagToStock(tagId) {
            if (!this.currentStockForTag) return;
            
            try {
                const response = await fetch(`${API_BASE}/stock-tags/${this.currentStockForTag.symbol}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ tag_id: tagId })
                });
                
                if (response.ok) {
                    this.showToast('标签添加成功', 'success');
                    await this.loadStockTags(this.currentStockForTag.symbol);
                } else {
                    this.showToast('标签添加失败', 'error');
                }
            } catch (error) {
                console.error('Failed to add tag:', error);
                this.showToast('标签添加失败', 'error');
            }
        },
        
        async removeTagFromStock(tagId) {
            if (!this.currentStockForTag) return;
            
            try {
                const response = await fetch(`${API_BASE}/stock-tags/${this.currentStockForTag.symbol}/${tagId}`, {
                    method: 'DELETE'
                });
                
                if (response.ok) {
                    this.showToast('标签移除成功', 'success');
                    await this.loadStockTags(this.currentStockForTag.symbol);
                } else {
                    this.showToast('标签移除失败', 'error');
                }
            } catch (error) {
                console.error('Failed to remove tag:', error);
                this.showToast('标签移除失败', 'error');
            }
        },
        
        async createAndAddTag() {
            if (!this.newTagName.trim() || !this.currentStockForTag) return;
            
            try {
                // First create the tag
                const createResponse = await fetch(`${API_BASE}/tags`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: this.newTagName.trim(),
                        color: this.newTagColor,
                        description: ''
                    })
                });
                
                if (!createResponse.ok) {
                    this.showToast('创建标签失败', 'error');
                    return;
                }
                
                const newTag = await createResponse.json();
                
                // Add the new tag to the stock
                await this.addTagToStock(newTag.id);
                
                // Reload all tags
                await this.loadTags();
                
                this.newTagName = '';
            } catch (error) {
                console.error('Failed to create and add tag:', error);
                this.showToast('创建标签失败', 'error');
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
                    // Clear sparkline cache when new filter is applied
                    this.loadedSparklines = {};
                }
            } catch (error) {
                console.error('Error filtering stocks:', error);
                this.showError('筛选失败: ' + error.message);
                this.filteredStocks = [];
            } finally {
                this.filterLoading = false;
            }
        },
        
        loadSparklineOnHover(stockCode) {
            console.log('loadSparklineOnHover called for:', stockCode);
            console.log('loadedSparklines:', this.loadedSparklines);
            
            // Only load once per stock
            if (this.loadedSparklines[stockCode]) {
                console.log('Already loaded, skipping');
                return;
            }
            
            this.loadedSparklines[stockCode] = true;
            console.log('Calling renderSparkline');
            this.renderSparkline(stockCode);
        },
        
        async renderSparkline(stockCode) {
            try {
                // Try to find container in either analytics or strategy tab
                let container = document.getElementById(`sparkline-${stockCode}`);
                if (!container) {
                    container = document.getElementById(`sparkline-strategy-${stockCode}`);
                }
                if (!container) {
                    return;
                }
                
                // Show loading state
                container.innerHTML = '<i class="fas fa-spinner fa-spin text-gray-400"></i>';
                
                // Fetch last 30 days of history
                const response = await fetch(`${API_BASE}/history/${stockCode}?limit=30`);
                if (!response.ok) {
                    container.innerHTML = '<span style="color: #ef4444; font-size: 10px;">错误</span>';
                    return;
                }
                
                const result = await response.json();
                
                if (result.error) {
                    container.innerHTML = '<span style="color: #ef4444; font-size: 10px;">错误</span>';
                    return;
                }
                
                if (!result.data || result.data.length === 0) {
                    container.innerHTML = '<span style="color: #999; font-size: 10px;">无数据</span>';
                    return;
                }
                
                const history = result.data.reverse(); // Oldest to newest
                const closes = history.map(h => h.close);
                
                // Check if echarts is available
                if (typeof echarts === 'undefined') {
                    container.innerHTML = '<span style="color: #ef4444; font-size: 10px;">错误</span>';
                    return;
                }
                
                // Clear container
                container.innerHTML = '';
                
                // Initialize mini chart
                const chart = echarts.init(container);
                
                // Determine color based on trend
                const firstClose = closes[0];
                const lastClose = closes[closes.length - 1];
                const isPositive = lastClose >= firstClose;
                const lineColor = isPositive ? '#10b981' : '#ef4444';
                const areaColor = isPositive ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)';
                
                const option = {
                    animation: false,
                    grid: {
                        left: 2,
                        right: 2,
                        top: 2,
                        bottom: 2
                    },
                    xAxis: {
                        type: 'category',
                        data: closes.map((_, i) => i),
                        show: false,
                        boundaryGap: false
                    },
                    yAxis: {
                        type: 'value',
                        show: false,
                        scale: true
                    },
                    series: [{
                        type: 'line',
                        data: closes,
                        smooth: true,
                        symbol: 'none',
                        lineStyle: {
                            color: lineColor,
                            width: 1.5
                        },
                        areaStyle: {
                            color: areaColor
                        }
                    }]
                };
                
                chart.setOption(option);
            } catch (error) {
                const container = document.getElementById(`sparkline-${stockCode}`);
                if (container) {
                    container.innerHTML = '<span style="color: #ef4444; font-size: 10px;">错误</span>';
                }
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
            if (!this.searchQuery.trim()) {
                this.loadStockList();
            }
        },
        
        prevPage() {
            if (this.currentPage > 1) {
                this.currentPage--;
                if (!this.searchQuery.trim()) {
                    this.loadStockList();
                }
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
            this.showToast(message, 'error');
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
            } else if (newTab === 'watchlist') {
                this.loadWatchlistStocks();
            } else if (newTab === 'warninglist') {
                this.loadWarninglistStocks();
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
        },
        
        searchQuery(newQuery) {
            // When user starts searching, load all stocks for client-side filtering
            if (newQuery.trim() && this.allStocks.length === 0) {
                this.loadStockList();
            }
            // Reset to page 1 when searching
            if (newQuery.trim()) {
                this.currentPage = 1;
            }
        },
        
        boardFiltersStock: {
            handler() {
                // When board filters change, clear cache and reload
                if (this.activeTab === 'stocks') {
                    this.allStocks = []; // Clear cache to force reload with new filters
                    this.currentPage = 1;
                    this.loadStockList();
                }
            },
            deep: true
        }
    }
}).mount('#app');
