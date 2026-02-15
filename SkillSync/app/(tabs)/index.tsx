import { StyleSheet, View, Text, TouchableOpacity, ScrollView } from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/theme';
import { useColorScheme } from '@/hooks/use-color-scheme';
import { useAuth } from '@/context/AuthContext';
import { IconSymbol } from '@/components/ui/icon-symbol';

export default function HomeScreen() {
  const colorScheme = useColorScheme();
  const theme = Colors[colorScheme ?? 'light'];
  const { user } = useAuth();
  const router = useRouter();

  const userEmail = user?.email ?? 'User';
  const firstName = userEmail.split('@')[0];

  return (
    <ScrollView
      style={[styles.container, { backgroundColor: theme.background }]}
      contentContainerStyle={styles.content}
    >
      {/* Header */}
      <View style={styles.headerSection}>
        <Text style={[styles.greeting, { color: theme.icon }]}>Welcome back,</Text>
        <Text style={[styles.name, { color: theme.text }]}>{firstName}</Text>
      </View>

      {/* Quick Action Cards */}
      <View style={styles.cardsSection}>
        <TouchableOpacity
          style={[styles.actionCard, { backgroundColor: theme.tint + '15', borderColor: theme.tint + '30' }]}
          onPress={() => router.push('/(tabs)/profile')}
          activeOpacity={0.7}
        >
          <View style={[styles.iconCircle, { backgroundColor: theme.tint + '25' }]}>
            <IconSymbol size={28} name="doc.text.fill" color={theme.tint} />
          </View>
          <Text style={[styles.cardTitle, { color: theme.text }]}>Upload Resume</Text>
          <Text style={[styles.cardDesc, { color: theme.icon }]}>
            Upload your resume to get AI-powered job matches
          </Text>
          <View style={styles.cardAction}>
            <Text style={[styles.cardActionText, { color: theme.tint }]}>Go to Profile</Text>
            <IconSymbol size={14} name="chevron.right" color={theme.tint} />
          </View>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.actionCard, { backgroundColor: '#22c55e15', borderColor: '#22c55e30' }]}
          onPress={() => router.push('/(tabs)/explore')}
          activeOpacity={0.7}
        >
          <View style={[styles.iconCircle, { backgroundColor: '#22c55e25' }]}>
            <IconSymbol size={28} name="paperplane.fill" color="#22c55e" />
          </View>
          <Text style={[styles.cardTitle, { color: theme.text }]}>View Matches</Text>
          <Text style={[styles.cardDesc, { color: theme.icon }]}>
            See jobs that match your skills and experience
          </Text>
          <View style={styles.cardAction}>
            <Text style={[styles.cardActionText, { color: '#22c55e' }]}>Explore Jobs</Text>
            <IconSymbol size={14} name="chevron.right" color="#22c55e" />
          </View>
        </TouchableOpacity>
      </View>

      {/* How It Works */}
      <View style={styles.howSection}>
        <Text style={[styles.sectionTitle, { color: theme.text }]}>How SkillSync Works</Text>

        {[
          { step: '1', icon: 'doc.text.fill' as const, title: 'Upload Resume', desc: 'Upload your resume (PDF or DOCX) on the Profile tab' },
          { step: '2', icon: 'gearshape.fill' as const, title: 'AI Analysis', desc: 'We extract your skills and create a smart profile' },
          { step: '3', icon: 'checkmark.circle.fill' as const, title: 'Get Matches', desc: 'View ranked job matches on the Explore tab' },
        ].map((item) => (
          <View key={item.step} style={[styles.stepRow, { borderBottomColor: theme.icon + '20' }]}>
            <View style={[styles.stepBadge, { backgroundColor: theme.tint }]}>
              <Text style={styles.stepNumber}>{item.step}</Text>
            </View>
            <View style={styles.stepContent}>
              <Text style={[styles.stepTitle, { color: theme.text }]}>{item.title}</Text>
              <Text style={[styles.stepDesc, { color: theme.icon }]}>{item.desc}</Text>
            </View>
          </View>
        ))}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    paddingTop: 60,
    paddingHorizontal: 20,
    paddingBottom: 32,
  },
  headerSection: {
    marginBottom: 28,
  },
  greeting: {
    fontSize: 16,
    fontWeight: '500',
  },
  name: {
    fontSize: 28,
    fontWeight: 'bold',
    marginTop: 2,
  },
  cardsSection: {
    gap: 14,
    marginBottom: 32,
  },
  actionCard: {
    borderWidth: 1,
    borderRadius: 16,
    padding: 20,
  },
  iconCircle: {
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 12,
  },
  cardTitle: {
    fontSize: 18,
    fontWeight: '700',
    marginBottom: 4,
  },
  cardDesc: {
    fontSize: 14,
    lineHeight: 20,
    marginBottom: 12,
  },
  cardAction: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  cardActionText: {
    fontSize: 14,
    fontWeight: '600',
  },
  howSection: {
    marginBottom: 20,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  stepRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    borderBottomWidth: 1,
    gap: 14,
  },
  stepBadge: {
    width: 32,
    height: 32,
    borderRadius: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  stepNumber: {
    color: '#fff',
    fontWeight: 'bold',
    fontSize: 14,
  },
  stepContent: {
    flex: 1,
  },
  stepTitle: {
    fontSize: 16,
    fontWeight: '600',
  },
  stepDesc: {
    fontSize: 13,
    marginTop: 2,
  },
});
