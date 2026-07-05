AppConfig.mounted = function () {
    this.loadWatchlist();
    this.init();
};

AppConfig.beforeUnmount = function () {
    if (this.dailyUpdateInterval) {
        clearInterval(this.dailyUpdateInterval);
    }
    if (this.backfillInterval) {
        clearInterval(this.backfillInterval);
    }
};

AppConfig.watch = {
    activeTab(newTab) {
        if (newTab === 'admin') {
            this.loadDailyUpdateStatus();
            this.loadBackfillStatus();
            this.loadBackfillGaps();
        } else if (newTab === 'gainers') {
            this.$nextTick(() => {
                if (this.topGainers.length > 0) {
                    this.renderGainersChart();
                } else {
                    this.loadTopGainers();
                }
            });
        } else if (newTab === 'losers') {
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
        if (this.activeTab === 'gainers') {
            this.loadTopGainers();
        } else if (this.activeTab === 'losers') {
            this.loadTopLosers();
        }
        this.loadStatsForDate(newDate);
    },

    searchQuery(newQuery) {
        if (newQuery.trim() && this.allStocks.length === 0) {
            this.loadStockList();
        }
        if (newQuery.trim()) {
            this.currentPage = 1;
        }
    },

    boardFiltersStock: {
        handler() {
            if (this.activeTab === 'stocks') {
                this.allStocks = [];
                this.currentPage = 1;
                this.loadStockList();
            }
        },
        deep: true
    }
};

Vue.createApp(AppConfig).mount('#app');
