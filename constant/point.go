package constant

var (
	RouteMap = map[uint8]string{
		1: "朝晖",
		2: "屏峰半程",
		3: "屏峰全程",
		4: "莫干山半程",
		5: "莫干山全程",
	}

	ZHMap = map[int8]string{
		0: "起点",
		1: "上塘映翠",
		2: "京杭大运河",
		3: "西湖文化广场",
		4: "中国海事",
		5: "忠亭",
		6: "德胜运河驿站",
		7: "终点",
	}

	PFHalfMap = map[int8]string{
		0: "起点",
		1: "金莲寺",
		2: "老焦山",
		3: "屏峰山",
		4: "屏峰善院",
		5: "终点",
	}

	PFAllMap = map[int8]string{
		0: "起点",
		1: "金莲寺",
		2: "白龙潭",
		3: "慈母桥",
		4: "古樟树公园",
		5: "屏峰山",
		6: "屏峰善院",
		7: "终点",
	}

	MgsHalfMap = map[int8]string{
		0: "起点",
		1: "终点",
	}

	MgsAllMap = map[int8]string{
		0: "起点",
		1: "兆丰公园",
		2: "滑板公园",
		3: "天安云谷",
		4: "东苕溪",
		5: "终点",
	}

	PointMap = map[uint8]uint8{
		1: 7,
		2: 5,
		3: 7,
		4: 1,
		5: 5,
	}
)

// GetPointName 用于根据 Route 和 Point 返回对应的点位名称
func GetPointName(route uint8, point int8) string {
	switch route {
	case 1:
		return ZHMap[point]
	case 2:
		return PFHalfMap[point]
	case 3:
		return PFAllMap[point]
	case 4:
		return MgsHalfMap[point]
	case 5:
		return MgsAllMap[point]
	default:
		return "未知点位"
	}
}
