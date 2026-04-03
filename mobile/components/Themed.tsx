import {
  Text as DefaultText,
  View as DefaultView,
  Pressable,
  PressableProps,
  StyleSheet,
} from 'react-native';

import { useColorScheme } from './useColorScheme';
import Colors from '@/constants/Colors';
import { typography, spacing, radius } from '@/constants/Typography';

type ThemeProps = {
  lightColor?: string;
  darkColor?: string;
};

export type TextProps = ThemeProps & DefaultText['props'];
export type ViewProps = ThemeProps & DefaultView['props'];

export function useThemeColor(
  props: { light?: string; dark?: string },
  colorName: keyof typeof Colors.light & keyof typeof Colors.dark
) {
  const theme = useColorScheme();
  const colorFromProps = props[theme];
  return colorFromProps ?? Colors[theme][colorName];
}

export function useColors() {
  const theme = useColorScheme();
  return Colors[theme];
}

export function Text(props: TextProps) {
  const { style, lightColor, darkColor, ...otherProps } = props;
  const color = useThemeColor({ light: lightColor, dark: darkColor }, 'text');
  return <DefaultText style={[{ color }, style]} {...otherProps} />;
}

export function View(props: ViewProps) {
  const { style, lightColor, darkColor, ...otherProps } = props;
  const backgroundColor = useThemeColor({ light: lightColor, dark: darkColor }, 'background');
  return <DefaultView style={[{ backgroundColor }, style]} {...otherProps} />;
}

export function Card({ style, children, ...props }: ViewProps) {
  const colors = useColors();
  return (
    <DefaultView
      style={[
        {
          backgroundColor: colors.card,
          borderRadius: radius.lg,
          borderWidth: 1,
          borderColor: colors.border,
          padding: spacing.md,
        },
        style,
      ]}
      {...props}
    >
      {children}
    </DefaultView>
  );
}

type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost';

type ButtonProps = PressableProps & {
  title: string;
  variant?: ButtonVariant;
  icon?: string;
  fullWidth?: boolean;
};

export function Button({ title, variant = 'primary', icon, fullWidth, style, ...props }: ButtonProps) {
  const colors = useColors();

  const bgColors: Record<ButtonVariant, string> = {
    primary: colors.tint,
    secondary: colors.tintLight,
    outline: 'transparent',
    ghost: 'transparent',
  };

  const textColors: Record<ButtonVariant, string> = {
    primary: '#FFFFFF',
    secondary: colors.tint,
    outline: colors.tint,
    ghost: colors.tint,
  };

  return (
    <Pressable
      style={({ pressed }) => [
        {
          backgroundColor: bgColors[variant],
          borderRadius: radius.md,
          paddingVertical: spacing.sm + 4,
          paddingHorizontal: spacing.lg,
          alignItems: 'center' as const,
          justifyContent: 'center' as const,
          flexDirection: 'row' as const,
          gap: spacing.sm,
          opacity: pressed ? 0.8 : 1,
          borderWidth: variant === 'outline' ? 1.5 : 0,
          borderColor: variant === 'outline' ? colors.tint : undefined,
          width: fullWidth ? '100%' : undefined,
        },
        style as any,
      ]}
      {...props}
    >
      {icon && <DefaultText style={{ fontSize: 18 }}>{icon}</DefaultText>}
      <DefaultText
        style={[
          typography.bodyBold,
          { color: textColors[variant] },
        ]}
      >
        {title}
      </DefaultText>
    </Pressable>
  );
}

type BadgeProps = {
  label: string;
  color?: string;
  backgroundColor?: string;
};

export function Badge({ label, color, backgroundColor }: BadgeProps) {
  const colors = useColors();
  return (
    <DefaultView
      style={{
        backgroundColor: backgroundColor ?? colors.tintLight,
        borderRadius: radius.full,
        paddingVertical: 2,
        paddingHorizontal: spacing.sm,
      }}
    >
      <DefaultText style={[typography.small, { color: color ?? colors.tint }]}>
        {label}
      </DefaultText>
    </DefaultView>
  );
}

export function ProgressBar({
  progress,
  color,
  height = 8,
  testID,
}: {
  progress: number;
  color?: string;
  height?: number;
  testID?: string;
}) {
  const colors = useColors();
  const clamp = Math.max(0, Math.min(1, progress));
  return (
    <DefaultView
      testID={testID}
      style={{
        height,
        borderRadius: height / 2,
        backgroundColor: colors.border,
        overflow: 'hidden',
      }}
    >
      <DefaultView
        style={{
          height: '100%',
          width: `${clamp * 100}%`,
          borderRadius: height / 2,
          backgroundColor: color ?? colors.tint,
        }}
      />
    </DefaultView>
  );
}

export function Separator() {
  const colors = useColors();
  return (
    <DefaultView
      style={{
        height: StyleSheet.hairlineWidth,
        backgroundColor: colors.border,
        marginVertical: spacing.md,
      }}
    />
  );
}
