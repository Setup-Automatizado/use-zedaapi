"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	ResponsiveContainer,
	LineChart,
	Line,
	XAxis,
	YAxis,
	Tooltip,
	CartesianGrid,
	PieChart,
	Pie,
	Cell,
} from "recharts";

const COLORS = [
	"hsl(var(--primary))",
	"hsl(217, 91%, 60%)",
	"hsl(142, 76%, 36%)",
	"hsl(280, 65%, 60%)",
	"hsl(24, 94%, 50%)",
];

interface AdminChartsClientProps {
	revenue: Array<{ month: string; value: number }>;
	plans: Array<{ name: string; value: number; fill?: string }>;
}

export function AdminChartsClient({ revenue, plans }: AdminChartsClientProps) {
	return (
		<div className="grid gap-4 lg:grid-cols-2">
			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-medium">
						Receita Mensal (MRR)
					</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="h-64">
						<ResponsiveContainer width="100%" height="100%">
							<LineChart data={revenue}>
								<CartesianGrid
									strokeDasharray="3 3"
									className="stroke-border"
								/>
								<XAxis
									dataKey="month"
									className="text-xs fill-muted-foreground"
								/>
								<YAxis
									className="text-xs fill-muted-foreground"
									tickFormatter={(v) =>
										`R$ ${(v / 1000).toFixed(0)}k`
									}
								/>
								<Tooltip
									formatter={(value: number) => [
										`R$ ${value.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}`,
										"Receita",
									]}
									contentStyle={{
										backgroundColor:
											"hsl(var(--popover))",
										borderColor:
											"hsl(var(--border))",
										borderRadius: "0.5rem",
										color: "hsl(var(--popover-foreground))",
									}}
								/>
								<Line
									type="monotone"
									dataKey="value"
									stroke="hsl(var(--primary))"
									strokeWidth={2}
									dot={{
										r: 4,
										fill: "hsl(var(--primary))",
									}}
								/>
							</LineChart>
						</ResponsiveContainer>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-medium">
						Assinaturas por Plano
					</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="h-64">
						<ResponsiveContainer width="100%" height="100%">
							<PieChart>
								<Pie
									data={plans}
									dataKey="value"
									nameKey="name"
									cx="50%"
									cy="50%"
									innerRadius={60}
									outerRadius={90}
								>
									{plans.map((_, i) => (
										<Cell
											key={i}
											fill={
												COLORS[i % COLORS.length]
											}
										/>
									))}
								</Pie>
								<Tooltip
									formatter={(
										value: number,
										name: string,
									) => [value, name]}
									contentStyle={{
										backgroundColor:
											"hsl(var(--popover))",
										borderColor:
											"hsl(var(--border))",
										borderRadius: "0.5rem",
										color: "hsl(var(--popover-foreground))",
									}}
								/>
							</PieChart>
						</ResponsiveContainer>
					</div>
					<div className="flex justify-center gap-6 text-xs">
						{plans.map((item, i) => (
							<div
								key={item.name}
								className="flex items-center gap-1.5"
							>
								<div
									className="size-2.5 rounded-full"
									style={{
										backgroundColor:
											COLORS[i % COLORS.length],
									}}
								/>
								<span className="text-muted-foreground">
									{item.name} ({item.value})
								</span>
							</div>
						))}
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
