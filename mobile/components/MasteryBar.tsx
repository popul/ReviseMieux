import { View as RNView, Text as RNText } from 'react-native';
import { masteryColors } from '@/constants/Colors';
import { useColors } from '@/components/Themed';
import { typography, spacing, radius } from '@/constants/Typography';

type MasteryBreakdown = {
  unknown: number;
  fragile: number;
  ok: number;
  solid: number;
};

type Props = {
  breakdown: MasteryBreakdown;
  height?: number;
  showLabel?: boolean;
};

export function MasteryBar({ breakdown, height = 10, showLabel = true }: Props) {
  const colors = useColors();
  const total = breakdown.unknown + breakdown.fragile + breakdown.ok + breakdown.solid;
  if (total === 0) return null;

  const mastered = breakdown.ok + breakdown.solid;
  const pct = Math.round((mastered / total) * 100);

  const segments = [
    { key: 'solid', value: breakdown.solid, color: masteryColors.solid },
    { key: 'ok', value: breakdown.ok, color: masteryColors.ok },
    { key: 'fragile', value: breakdown.fragile, color: masteryColors.fragile },
    { key: 'unknown', value: breakdown.unknown, color: masteryColors.unknown },
  ].filter((s) => s.value > 0);

  return (
    <RNView style={{ gap: spacing.xs }}>
      <RNView
        style={{
          height,
          borderRadius: height / 2,
          backgroundColor: colors.border,
          overflow: 'hidden',
          flexDirection: 'row',
        }}
      >
        {segments.map((seg) => (
          <RNView
            key={seg.key}
            style={{
              flex: seg.value,
              backgroundColor: seg.color,
            }}
          />
        ))}
      </RNView>
      {showLabel && (
        <RNText style={[typography.small, { color: colors.textSecondary }]}>
          {pct}% maitrise
        </RNText>
      )}
    </RNView>
  );
}

type MasteryDotProps = {
  state: 'unknown' | 'fragile' | 'ok' | 'solid';
};

export function MasteryDot({ state }: MasteryDotProps) {
  return (
    <RNView
      style={{
        width: 10,
        height: 10,
        borderRadius: 5,
        backgroundColor: masteryColors[state],
      }}
    />
  );
}

const stateLabels: Record<string, string> = {
  unknown: 'NEW',
  fragile: 'FRAGILE',
  ok: 'OK',
  solid: 'SOLID',
};

export function MasteryBadge({ state }: MasteryDotProps) {
  const bg = masteryColors[state];
  return (
    <RNView
      style={{
        backgroundColor: bg + '20',
        borderRadius: radius.full,
        paddingVertical: 2,
        paddingHorizontal: spacing.sm,
        flexDirection: 'row',
        alignItems: 'center',
        gap: 4,
      }}
    >
      <MasteryDot state={state} />
      <RNText style={[typography.small, { color: bg, fontWeight: '600' }]}>
        {stateLabels[state]}
      </RNText>
    </RNView>
  );
}
