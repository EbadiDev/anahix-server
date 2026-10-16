package admin

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/jalali"
)

type BIAnalyticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBIAnalyticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BIAnalyticsLogic {
	return &BIAnalyticsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetOverview compiles end-to-end commercial BI metrics, graphs, and KPIs
func (l *BIAnalyticsLogic) GetOverview(req *types.AdminBIOverviewRequest) (*types.AdminBIOverviewResponse, error) {
	effectiveUSDRate := l.getEffectiveUSDRate()

	// 1. Determine time boundaries
	now := time.Now()
	startTime := now.AddDate(0, 0, -30) // default 30 days
	endTime := now

	rangeStr := strings.ToLower(strings.TrimSpace(req.Range))
	switch rangeStr {
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "30d":
		startTime = now.AddDate(0, 0, -30)
	case "90d":
		startTime = now.AddDate(0, 0, -90)
	case "1y":
		startTime = now.AddDate(-1, 0, 0)
	case "all":
		startTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		if req.StartDate != "" {
			if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
				startTime = t
			}
		}
		if req.EndDate != "" {
			if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
				endTime = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			}
		}
	}

	// 2. Fetch Store Orders in range
	var allOrders []model.Order
	l.svcCtx.DB.WithContext(l.ctx).
		Preload("Items").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Find(&allOrders)

	// Pre-load all products to have cost fallback map
	var products []model.Product
	l.svcCtx.DB.WithContext(l.ctx).Find(&products)
	prodCostMap := make(map[string]float64)
	prodTitleMap := make(map[string]string)
	prodCatMap := make(map[string]string)
	prodSlugMap := make(map[string]string)
	for _, p := range products {
		prodCostMap[p.ID.String()] = p.BaseCostUSD
		prodTitleMap[p.ID.String()] = p.TitleFA
		prodCatMap[p.ID.String()] = string(p.Category)
		prodSlugMap[p.ID.String()] = p.Slug
	}

	// 3. Fetch Virtual Number Orders in range
	var vnOrders []model.VirtualNumberOrder
	l.svcCtx.DB.WithContext(l.ctx).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Find(&vnOrders)

	// Aggregators
	var totalOrdersCount int64 = int64(len(allOrders))
	var paidOrdersCount int64 = 0
	var storeProductsRevenue int64 = 0
	var storeProductsCOGS int64 = 0
	var storeProductsCount int64 = 0
	var vnRevenue int64 = 0
	var vnCOGS int64 = 0
	var vnCount int64 = 0

	type prodSaleStat struct {
		id       string
		title    string
		slug     string
		cat      string
		units    int64
		revenue  int64
		cogs     int64
	}
	productSalesMap := make(map[string]*prodSaleStat)

	uniqueCustomers := make(map[string]struct{})

	// Time series grouping map: key = period bucket string
	type timeBucket struct {
		dateStr      string
		periodLabel  string
		timestamp    time.Time
		revenueToman int64
		cogsToman    int64
		totalOrders  int64
		paidOrders   int64
	}
	buckets := make(map[string]*timeBucket)

	getBucketKey := func(t time.Time) (string, string) {
		loc, _ := time.LoadLocation("Asia/Tehran")
		if loc != nil {
			t = t.In(loc)
		}
		if req.GroupBy == "monthly" {
			key := t.Format("2006-01")
			jy, jm, _ := jalali.ToJalali(t)
			label := fmt.Sprintf("%04d/%02d", jy, jm)
			return key, label
		} else if req.GroupBy == "weekly" {
			year, week := t.ISOWeek()
			key := fmt.Sprintf("%d-W%02d", year, week)
			label := fmt.Sprintf("هفته %d", week)
			return key, label
		}
		// default daily
		key := t.Format("2006-01-02")
		label := jalali.FormatDateJalali(t)
		return key, label
	}

	for _, o := range allOrders {
		bKey, bLabel := getBucketKey(o.CreatedAt)
		if _, exists := buckets[bKey]; !exists {
			buckets[bKey] = &timeBucket{
				dateStr:     o.CreatedAt.Format("2006-01-02"),
				periodLabel: bLabel,
				timestamp:   o.CreatedAt,
			}
		}
		b := buckets[bKey]
		b.totalOrders++

		isOrderPaid := o.Status == model.OrderPaid ||
			o.Status == model.OrderCompleted ||
			o.Status == model.OrderInQueue ||
			o.Status == model.OrderProcessing ||
			o.Status == model.OrderAwaitingCustomerAction

		if isOrderPaid {
			paidOrdersCount++
			b.paidOrders++
			storeProductsRevenue += o.TotalAmountToman
			b.revenueToman += o.TotalAmountToman

			if o.CustomerPhone != "" {
				uniqueCustomers[o.CustomerPhone] = struct{}{}
			}

			// Calculate COGS per item
			for _, item := range o.Items {
				storeProductsCount += int64(item.Quantity)
				pID := item.ProductID.String()

				itemCOGSToman := item.CostToman * int64(item.Quantity)
				if itemCOGSToman <= 0 {
					baseCostUSD := item.CostUSD
					if baseCostUSD <= 0 {
						baseCostUSD = prodCostMap[pID]
					}
					itemCOGSToman = int64(math.Round(baseCostUSD * effectiveUSDRate * float64(item.Quantity)))
				}

				storeProductsCOGS += itemCOGSToman
				b.cogsToman += itemCOGSToman

				// Track top product stats
				stat, ok := productSalesMap[pID]
				if !ok {
					stat = &prodSaleStat{
						id:    pID,
						title: item.ProductTitle,
						slug:  item.ProductSlug,
						cat:   prodCatMap[pID],
					}
					if stat.title == "" {
						stat.title = prodTitleMap[pID]
					}
					if stat.slug == "" {
						stat.slug = prodSlugMap[pID]
					}
					productSalesMap[pID] = stat
				}
				stat.units += int64(item.Quantity)
				stat.revenue += item.SubtotalToman
				stat.cogs += itemCOGSToman
			}
		}
	}

	// Process Virtual Numbers (5SIM)
	type countryStat struct {
		country    string
		countryFA  string
		flag       string
		count      int64
		revenue    int64
		serviceMap map[string]int64
	}
	countriesMap := make(map[string]*countryStat)

	for _, vn := range vnOrders {
		// Only count received / finished / paid virtual numbers
		if vn.Status == model.VNumStatusReceived || vn.Status == model.VNumStatusFinished {
			vnCount++
			vnRevenue += vn.PriceToman
			
			// COGS for virtual numbers
			vnCostToman := int64(math.Round(vn.BaseCost * effectiveUSDRate))
			if vnCostToman <= 0 && vn.BaseCost > 0 {
				vnCostToman = int64(vn.BaseCost * effectiveUSDRate)
			}
			vnCOGS += vnCostToman

			if vn.CustomerPhone != "" {
				uniqueCustomers[vn.CustomerPhone] = struct{}{}
			}

			// Add to bucket
			bKey, bLabel := getBucketKey(vn.CreatedAt)
			if _, exists := buckets[bKey]; !exists {
				buckets[bKey] = &timeBucket{
					dateStr:     vn.CreatedAt.Format("2006-01-02"),
					periodLabel: bLabel,
					timestamp:   vn.CreatedAt,
				}
			}
			b := buckets[bKey]
			b.revenueToman += vn.PriceToman
			b.cogsToman += vnCostToman
			b.paidOrders++
			b.totalOrders++

			// Track country stats
			cKey := vn.Country
			if cKey == "" {
				cKey = "unknown"
			}
			cStat, exists := countriesMap[cKey]
			if !exists {
				cStat = &countryStat{
					country:    cKey,
					countryFA:  vn.CountryNameFA,
					flag:       vn.CountryFlag,
					serviceMap: make(map[string]int64),
				}
				if cStat.countryFA == "" {
					cStat.countryFA = vn.Country
				}
				countriesMap[cKey] = cStat
			}
			cStat.count++
			cStat.revenue += vn.PriceToman
			cStat.serviceMap[vn.ServiceNameFA]++
		}
	}

	// 4. Compute Totals & Headline KPIs
	totalRevenueToman := storeProductsRevenue + vnRevenue
	totalCOGSToman := storeProductsCOGS + vnCOGS
	grossProfitToman := totalRevenueToman - totalCOGSToman
	var grossMarginPercent float64 = 0
	if totalRevenueToman > 0 {
		grossMarginPercent = math.Round((float64(grossProfitToman)/float64(totalRevenueToman))*1000) / 10
	}

	totalEffectivePaidOrders := paidOrdersCount + vnCount
	totalEffectiveOrders := totalOrdersCount + int64(len(vnOrders))

	var conversionRatePercent float64 = 0
	if totalEffectiveOrders > 0 {
		conversionRatePercent = math.Round((float64(totalEffectivePaidOrders)/float64(totalEffectiveOrders))*1000) / 10
	}

	var aovToman int64 = 0
	if totalEffectivePaidOrders > 0 {
		aovToman = totalRevenueToman / totalEffectivePaidOrders
	}

	toUSDT := func(toman int64) float64 {
		if effectiveUSDRate <= 0 {
			return 0
		}
		return math.Round((float64(toman)/effectiveUSDRate)*100) / 100
	}

	kpis := types.AdminBIKPIs{
		TotalRevenueToman:      totalRevenueToman,
		TotalRevenueUSDT:       toUSDT(totalRevenueToman),
		TotalCOGSToman:         totalCOGSToman,
		TotalCOGSUSDT:          toUSDT(totalCOGSToman),
		GrossProfitToman:       grossProfitToman,
		GrossProfitUSDT:        toUSDT(grossProfitToman),
		GrossMarginPercent:     grossMarginPercent,
		TotalOrders:            totalEffectiveOrders,
		PaidOrders:             totalEffectivePaidOrders,
		ConversionRatePercent:  conversionRatePercent,
		AverageOrderValueToman: aovToman,
		AverageOrderValueUSDT:  toUSDT(aovToman),
		TotalCustomers:         int64(len(uniqueCustomers)),
		VirtualNumbersRevenue:  vnRevenue,
		VirtualNumbersCount:    vnCount,
		StoreProductsRevenue:   storeProductsRevenue,
		StoreProductsCount:     storeProductsCount,
	}

	// 5. Build Sorted Time Series List
	var timeSeries []types.AdminBITimeSeriesPoint
	for _, b := range buckets {
		profit := b.revenueToman - b.cogsToman
		var margin float64 = 0
		if b.revenueToman > 0 {
			margin = math.Round((float64(profit)/float64(b.revenueToman))*1000) / 10
		}
		var convRate float64 = 0
		if b.totalOrders > 0 {
			convRate = math.Round((float64(b.paidOrders)/float64(b.totalOrders))*1000) / 10
		}

		timeSeries = append(timeSeries, types.AdminBITimeSeriesPoint{
			PeriodLabel:      b.periodLabel,
			Date:             b.dateStr,
			RevenueToman:     b.revenueToman,
			RevenueUSDT:      toUSDT(b.revenueToman),
			COGSToman:        b.cogsToman,
			COGSUSDT:         toUSDT(b.cogsToman),
			GrossProfitToman: profit,
			GrossProfitUSDT:  toUSDT(profit),
			GrossMargin:      margin,
			TotalOrders:      b.totalOrders,
			PaidOrders:       b.paidOrders,
			ConversionRate:   convRate,
		})
	}
	sort.Slice(timeSeries, func(i, j int) bool {
		return timeSeries[i].Date < timeSeries[j].Date
	})

	// 6. Build Top Products Ranking
	var topProducts []types.AdminBITopProduct
	for _, stat := range productSalesMap {
		profit := stat.revenue - stat.cogs
		var margin float64 = 0
		if stat.revenue > 0 {
			margin = math.Round((float64(profit)/float64(stat.revenue))*1000) / 10
		}
		topProducts = append(topProducts, types.AdminBITopProduct{
			ProductID:        stat.id,
			TitleFA:          stat.title,
			Slug:             stat.slug,
			Category:         stat.cat,
			UnitsSold:        stat.units,
			RevenueToman:     stat.revenue,
			COGSToman:        stat.cogs,
			GrossProfitToman: profit,
			MarginPercent:    margin,
		})
	}
	sort.Slice(topProducts, func(i, j int) bool {
		return topProducts[i].RevenueToman > topProducts[j].RevenueToman
	})
	if len(topProducts) > 10 {
		topProducts = topProducts[:10]
	}

	// 7. Build Top Countries Ranking for Virtual Numbers
	var topCountries []types.AdminBITopCountry
	for _, c := range countriesMap {
		bestService := ""
		var maxServiceCount int64 = 0
		for sName, sCount := range c.serviceMap {
			if sCount > maxServiceCount {
				maxServiceCount = sCount
				bestService = sName
			}
		}
		topCountries = append(topCountries, types.AdminBITopCountry{
			Country:       c.country,
			CountryNameFA: c.countryFA,
			CountryFlag:   c.flag,
			OrdersCount:   c.count,
			RevenueToman:  c.revenue,
			TopService:    bestService,
		})
	}
	sort.Slice(topCountries, func(i, j int) bool {
		return topCountries[i].OrdersCount > topCountries[j].OrdersCount
	})
	if len(topCountries) > 10 {
		topCountries = topCountries[:10]
	}

	// 8. Build Loyalty Tiers Breakdown & Customer Distribution
	loyaltyStats := l.calculateLoyaltyDistribution()

	return &types.AdminBIOverviewResponse{
		KPIs:             kpis,
		TimeSeries:       timeSeries,
		TopProducts:      topProducts,
		TopCountries:     topCountries,
		LoyaltyStats:     loyaltyStats,
		EffectiveUSDRate: effectiveUSDRate,
		GeneratedAt:      now.Format(time.RFC3339),
	}, nil
}

// getEffectiveUSDRate resolves current live or manual USD exchange rate
func (l *BIAnalyticsLogic) getEffectiveUSDRate() float64 {
	var setting model.SystemSetting
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("category = ? AND key = ?", "fivesim", "fivesim_usd_rate").
		First(&setting).Error
	if err == nil && setting.Value != "" {
		if rate, parseErr := strconv.ParseFloat(setting.Value, 64); parseErr == nil && rate > 0 {
			return rate
		}
	}

	// Fallback to general settings or default 95,000 Tomans
	return 95000.0
}

// calculateLoyaltyDistribution calculates customer count and spend per VIP tier
func (l *BIAnalyticsLogic) calculateLoyaltyDistribution() []types.AdminBILoyaltyDistribution {
	var tiers []model.LoyaltyTier
	l.svcCtx.DB.WithContext(l.ctx).Where("is_active = ?", true).Order("order_index asc").Find(&tiers)

	var totalCustomerCount int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Count(&totalCustomerCount)

	type TierSpendAgg struct {
		Tier       string
		Count      int64
		TotalSpent int64
	}
	var aggResults []TierSpendAgg
	l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Customer{}).
		Select("tier, count(*) as count, COALESCE(sum(total_spent_toman), 0) as total_spent").
		Group("tier").
		Scan(&aggResults)

	tierMap := make(map[string]TierSpendAgg)
	var grandTotalSpent int64 = 0
	for _, a := range aggResults {
		tierMap[strings.ToLower(a.Tier)] = a
		grandTotalSpent += a.TotalSpent
	}

	var res []types.AdminBILoyaltyDistribution
	for _, t := range tiers {
		agg := tierMap[strings.ToLower(t.Code)]
		var pctUsers float64 = 0
		if totalCustomerCount > 0 {
			pctUsers = math.Round((float64(agg.Count)/float64(totalCustomerCount))*1000) / 10
		}
		var pctSales float64 = 0
		if grandTotalSpent > 0 {
			pctSales = math.Round((float64(agg.TotalSpent)/float64(grandTotalSpent))*1000) / 10
		}

		res = append(res, types.AdminBILoyaltyDistribution{
			TierCode:          t.Code,
			TierName:          t.Name,
			TierNameEN:        t.NameEN,
			BadgeColor:        t.BadgeColor,
			DiscountPercent:   t.DiscountPercent,
			CustomerCount:     agg.Count,
			TotalSpentToman:   agg.TotalSpent,
			PercentageOfUsers: pctUsers,
			PercentageOfSales: pctSales,
		})
	}

	return res
}
