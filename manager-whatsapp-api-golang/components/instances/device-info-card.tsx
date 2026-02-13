"use client";

import { Building2, Info, Smartphone } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import type { DeviceInfo } from "@/types";

export interface DeviceInfoCardProps {
	deviceInfo: DeviceInfo;
	className?: string;
}

export function DeviceInfoCard({ deviceInfo, className }: DeviceInfoCardProps) {
	const { phone, name, imgUrl, device, isBusiness } = deviceInfo;

	const formatPhoneNumber = (phoneNumber: string) => {
		// Format phone number for display (WhatsApp style)
		const cleaned = phoneNumber.replace(/\D/g, "");

		// Brazilian: +55 DD NNNNN-NNNN (mobile) or +55 DD NNNN-NNNN (landline)
		if (cleaned.startsWith("55") && cleaned.length >= 12) {
			const areaCode = cleaned.substring(2, 4);
			const localNumber = cleaned.substring(4);

			if (localNumber.length === 9) {
				// Mobile: 9XXXX-XXXX
				const firstPart = localNumber.substring(0, 5);
				const secondPart = localNumber.substring(5);
				return `+55 ${areaCode} ${firstPart}-${secondPart}`;
			} else if (localNumber.length === 8) {
				// Landline: XXXX-XXXX
				const firstPart = localNumber.substring(0, 4);
				const secondPart = localNumber.substring(4);
				return `+55 ${areaCode} ${firstPart}-${secondPart}`;
			}
		}

		// Argentina mobile: +54 9 XXXX XX-XXXX (with 9 prefix)
		if (cleaned.startsWith("549") && cleaned.length >= 13) {
			const areaCode = cleaned.substring(3, 7);
			const localNumber = cleaned.substring(7);
			if (localNumber.length >= 6) {
				const firstPart = localNumber.substring(0, 2);
				const secondPart = localNumber.substring(2);
				return `+54 9 ${areaCode} ${firstPart}-${secondPart}`;
			}
		}

		// Argentina landline: +54 XXXX XX-XXXX
		if (
			cleaned.startsWith("54") &&
			!cleaned.startsWith("549") &&
			cleaned.length >= 12
		) {
			const areaCode = cleaned.substring(2, 6);
			const localNumber = cleaned.substring(6);
			if (localNumber.length >= 6) {
				const firstPart = localNumber.substring(0, 2);
				const secondPart = localNumber.substring(2);
				return `+54 ${areaCode} ${firstPart}-${secondPart}`;
			}
		}

		// USA/Canada: +1 XXX XXX-XXXX
		if (cleaned.startsWith("1") && cleaned.length === 11) {
			const areaCode = cleaned.substring(1, 4);
			const firstPart = cleaned.substring(4, 7);
			const secondPart = cleaned.substring(7);
			return `+1 ${areaCode} ${firstPart}-${secondPart}`;
		}

		// Colombia: +57 XXX XXX XXXX
		if (cleaned.startsWith("57") && cleaned.length >= 12) {
			const part1 = cleaned.substring(2, 5);
			const part2 = cleaned.substring(5, 8);
			const part3 = cleaned.substring(8);
			return `+57 ${part1} ${part2} ${part3}`;
		}

		// Mexico mobile: +52 1 XXX XXX XXXX (with 1 prefix)
		if (cleaned.startsWith("521") && cleaned.length >= 13) {
			const areaCode = cleaned.substring(3, 6);
			const firstPart = cleaned.substring(6, 9);
			const secondPart = cleaned.substring(9);
			return `+52 1 ${areaCode} ${firstPart} ${secondPart}`;
		}

		// Mexico landline: +52 XXX XXX XXXX
		if (
			cleaned.startsWith("52") &&
			!cleaned.startsWith("521") &&
			cleaned.length >= 12
		) {
			const areaCode = cleaned.substring(2, 5);
			const firstPart = cleaned.substring(5, 8);
			const secondPart = cleaned.substring(8);
			return `+52 ${areaCode} ${firstPart} ${secondPart}`;
		}

		// Spain: +34 XXX XX XX XX
		if (cleaned.startsWith("34") && cleaned.length >= 11) {
			const p1 = cleaned.substring(2, 5);
			const p2 = cleaned.substring(5, 7);
			const p3 = cleaned.substring(7, 9);
			const p4 = cleaned.substring(9);
			return `+34 ${p1} ${p2} ${p3} ${p4}`;
		}

		// Portugal: +351 XXX XXX XXX
		if (cleaned.startsWith("351") && cleaned.length >= 12) {
			const p1 = cleaned.substring(3, 6);
			const p2 = cleaned.substring(6, 9);
			const p3 = cleaned.substring(9);
			return `+351 ${p1} ${p2} ${p3}`;
		}

		// Chile: +56 X XXXX XXXX
		if (cleaned.startsWith("56") && cleaned.length >= 11) {
			const prefix = cleaned.substring(2, 3);
			const p1 = cleaned.substring(3, 7);
			const p2 = cleaned.substring(7);
			return `+56 ${prefix} ${p1} ${p2}`;
		}

		// Peru: +51 XXX XXX XXX
		if (cleaned.startsWith("51") && cleaned.length >= 11) {
			const p1 = cleaned.substring(2, 5);
			const p2 = cleaned.substring(5, 8);
			const p3 = cleaned.substring(8);
			return `+51 ${p1} ${p2} ${p3}`;
		}

		// Uruguay: +598 XX XXX XXX
		if (cleaned.startsWith("598") && cleaned.length >= 11) {
			const p1 = cleaned.substring(3, 5);
			const p2 = cleaned.substring(5, 8);
			const p3 = cleaned.substring(8);
			return `+598 ${p1} ${p2} ${p3}`;
		}

		// Paraguay: +595 XXX XXX XXX
		if (cleaned.startsWith("595") && cleaned.length >= 12) {
			const p1 = cleaned.substring(3, 6);
			const p2 = cleaned.substring(6, 9);
			const p3 = cleaned.substring(9);
			return `+595 ${p1} ${p2} ${p3}`;
		}

		// Bolivia: +591 X XXX XXXX
		if (cleaned.startsWith("591") && cleaned.length >= 11) {
			const prefix = cleaned.substring(3, 4);
			const p1 = cleaned.substring(4, 7);
			const p2 = cleaned.substring(7);
			return `+591 ${prefix} ${p1} ${p2}`;
		}

		// Ecuador: +593 XX XXX XXXX
		if (cleaned.startsWith("593") && cleaned.length >= 12) {
			const p1 = cleaned.substring(3, 5);
			const p2 = cleaned.substring(5, 8);
			const p3 = cleaned.substring(8);
			return `+593 ${p1} ${p2} ${p3}`;
		}

		// Venezuela: +58 XXX XXX XXXX
		if (cleaned.startsWith("58") && cleaned.length >= 12) {
			const p1 = cleaned.substring(2, 5);
			const p2 = cleaned.substring(5, 8);
			const p3 = cleaned.substring(8);
			return `+58 ${p1} ${p2} ${p3}`;
		}

		// UK: +44 XXXX XXXXXX
		if (cleaned.startsWith("44") && cleaned.length >= 12) {
			const p1 = cleaned.substring(2, 6);
			const p2 = cleaned.substring(6);
			return `+44 ${p1} ${p2}`;
		}

		// Germany: +49 XXX XXXXXXXX
		if (cleaned.startsWith("49") && cleaned.length >= 12) {
			const p1 = cleaned.substring(2, 5);
			const p2 = cleaned.substring(5);
			return `+49 ${p1} ${p2}`;
		}

		// France: +33 X XX XX XX XX
		if (cleaned.startsWith("33") && cleaned.length >= 11) {
			const p1 = cleaned.substring(2, 3);
			const p2 = cleaned.substring(3, 5);
			const p3 = cleaned.substring(5, 7);
			const p4 = cleaned.substring(7, 9);
			const p5 = cleaned.substring(9);
			return `+33 ${p1} ${p2} ${p3} ${p4} ${p5}`;
		}

		// Italy: +39 XXX XXX XXXX
		if (cleaned.startsWith("39") && cleaned.length >= 12) {
			const p1 = cleaned.substring(2, 5);
			const p2 = cleaned.substring(5, 8);
			const p3 = cleaned.substring(8);
			return `+39 ${p1} ${p2} ${p3}`;
		}

		// Generic international format fallback
		if (cleaned.length >= 10) {
			return `+${cleaned.substring(0, 2)} ${cleaned.substring(2)}`;
		}

		return phoneNumber;
	};

	return (
		<Card className={className}>
			<CardHeader>
				<CardTitle className="flex items-center gap-2">
					<Smartphone className="h-5 w-5" />
					Device Information
				</CardTitle>
			</CardHeader>
			<CardContent className="space-y-6">
				{/* Profile Section */}
				<div className="flex items-start gap-4">
					<Avatar className="h-16 w-16">
						<AvatarImage src={imgUrl} alt={name} />
						<AvatarFallback className="text-lg">
							{name
								.split(" ")
								.slice(0, 2)
								.map((n) => n[0])
								.join("")
								.toUpperCase()}
						</AvatarFallback>
					</Avatar>

					<div className="flex-1 space-y-1">
						<div className="flex items-center gap-2">
							<h3 className="font-semibold">{name}</h3>
							{isBusiness && (
								<Badge variant="outline" className="gap-1">
									<Building2 className="h-3 w-3" />
									Business
								</Badge>
							)}
						</div>
						<p className="text-sm text-muted-foreground">
							{formatPhoneNumber(phone)}
						</p>
					</div>
				</div>

				<Separator />

				{/* Device Details */}
				<div className="space-y-3">
					<h4 className="flex items-center gap-2 text-sm font-medium">
						<Info className="h-4 w-4" />
						Technical Details
					</h4>

					<div className="grid gap-3 text-sm">
						<div className="flex justify-between">
							<span className="text-muted-foreground">Model:</span>
							<span className="font-medium">
								{device.device_model || "N/A"}
							</span>
						</div>

						<div className="flex justify-between">
							<span className="text-muted-foreground">Platform:</span>
							<span className="font-medium capitalize">
								{device.platform || "N/A"}
							</span>
						</div>

						<div className="flex justify-between">
							<span className="text-muted-foreground">WhatsApp Version:</span>
							<span className="font-medium">{device.wa_version || "N/A"}</span>
						</div>

						<div className="flex justify-between">
							<span className="text-muted-foreground">OS Version:</span>
							<span className="font-medium">{device.os_version || "N/A"}</span>
						</div>

						{device.device_manufacturer && (
							<div className="flex justify-between">
								<span className="text-muted-foreground">Manufacturer:</span>
								<span className="font-medium">
									{device.device_manufacturer}
								</span>
							</div>
						)}

						{device.sessionName && (
							<div className="flex justify-between">
								<span className="text-muted-foreground">Session:</span>
								<span className="font-mono text-xs">{device.sessionName}</span>
							</div>
						)}
					</div>
				</div>

				{/* Additional Info */}
				{(device.mcc || device.mnc) && (
					<>
						<Separator />
						<div className="space-y-2">
							<h4 className="text-sm font-medium text-muted-foreground">
								Network Information
							</h4>
							<div className="grid gap-2 text-sm">
								{device.mcc && (
									<div className="flex justify-between">
										<span className="text-muted-foreground">MCC:</span>
										<span className="font-medium">{device.mcc}</span>
									</div>
								)}
								{device.mnc && (
									<div className="flex justify-between">
										<span className="text-muted-foreground">MNC:</span>
										<span className="font-medium">{device.mnc}</span>
									</div>
								)}
							</div>
						</div>
					</>
				)}
			</CardContent>
		</Card>
	);
}
