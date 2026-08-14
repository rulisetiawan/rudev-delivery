import React, { useState } from 'react';
import { StyleSheet, Text, View, TouchableOpacity, ScrollView, Alert, Image, TextInput, Linking } from 'react-native';
import * as ImagePicker from 'expo-image-picker';

export default function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [activeTab, setActiveTab] = useState('jobs'); // 'jobs' | 'active_job'
  const [selectedJob, setSelectedJob] = useState(null);
  const [receiptImage, setReceiptImage] = useState(null);
  const [actualPrice, setActualPrice] = useState('23000');

  const jobsList = [
    {
      id: 8821,
      code: 'ORD-8821',
      storeName: 'Warung Bu Ani',
      customerName: 'Pak Agus',
      customerPhone: '6281234567890',
      address: 'RT 03/RW 01 (Rumah Cat Hijau)',
      estimatedTotal: 33000,
      fee: 10000,
    }
  ];

  const handleClaimJob = (job) => {
    setSelectedJob(job);
    setActiveTab('active_job');
    Alert.alert('Job Berhasil Diambil!', `Silakan pergi ke ${job.storeName} untuk membeli barang.`);
  };

  const handlePickReceiptImage = async () => {
    const permissionResult = await ImagePicker.requestCameraPermissionsAsync();
    if (!permissionResult.granted) {
      Alert.alert('Izin Ditolak', 'Izin kamera dibutuhkan untuk mengambil foto nota warung!');
      return;
    }

    const result = await ImagePicker.launchCameraAsync({
      quality: 0.7,
    });

    if (!result.canceled) {
      setReceiptImage(result.assets[0].uri);
    }
  };

  const handleSubmitReceipt = () => {
    if (!receiptImage) {
      Alert.alert('Foto Nota Wajib', 'Silakan foto nota kertas warung terlebih dahulu!');
      return;
    }

    const totalCOD = parseFloat(actualPrice) + selectedJob.fee;

    Alert.alert(
      'Nota Berhasil Disimpan!',
      `Status pesanan diubah ke Dalam Pengantaran.\nTotal COD ditagih ke pembeli: Rp ${totalCOD.toLocaleString('id-ID')}`,
      [
        {
          text: 'Antar Sekarang',
          onPress: () => {
            const mapsUrl = `https://maps.google.com/?q=${encodeURIComponent(selectedJob.address)}`;
            Linking.openURL(mapsUrl);
          }
        }
      ]
    );
  };

  if (!isLoggedIn) {
    return (
      <View style={styles.centerContainer}>
        <Text style={styles.mainTitle}>🛵 DESA DRIVER APP</Text>
        <Text style={styles.subTitle}>Powered by Expo & Gluestack UI Components</Text>

        <TextInput style={styles.input} placeholder="Nomor HP Driver (Format 628xxx)" placeholderTextColor="#94a3b8" />
        <TouchableOpacity style={styles.btnPrimary} onPress={() => setIsLoggedIn(true)}>
          <Text style={styles.btnText}>MASUK / NARIK DRIVER</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>🛵 DESA DRIVER</Text>
        <Text style={styles.headerStatus}>🟢 STATUS: ONLINE / SIAP NARIK</Text>
      </View>

      <ScrollView style={styles.body}>
        {activeTab === 'jobs' ? (
          <View>
            <Text style={styles.sectionTitle}>📢 DAFTAR JOB TERSEDIA (1)</Text>

            {jobsList.map((job) => (
              <View key={job.id} style={styles.card}>
                <Text style={styles.cardCode}>#{job.code}</Text>
                <Text style={styles.cardStore}>🏪 {job.storeName}</Text>
                <Text style={styles.cardText}>👤 Pembeli: {job.customerName}</Text>
                <Text style={styles.cardText}>📍 Tujuan: {job.address}</Text>
                <Text style={styles.cardFee}>💰 Ongkir Hak Driver: Rp {job.fee.toLocaleString('id-ID')}</Text>

                <TouchableOpacity style={styles.btnClaim} onPress={() => handleClaimJob(job)}>
                  <Text style={styles.btnText}>AMBIL ORDERAN INI</Text>
                </TouchableOpacity>
              </View>
            ))}
          </View>
        ) : (
          <View>
            <Text style={styles.sectionTitle}>🚚 TUGAS PENGANTARAN AKTIF (#{selectedJob?.code})</Text>

            <View style={styles.card}>
              <Text style={styles.cardStore}>🏪 Jemput di: {selectedJob?.storeName}</Text>
              <Text style={styles.cardText}>👤 Antar ke: {selectedJob?.customerName} ({selectedJob?.customerPhone})</Text>
              <Text style={styles.cardText}>📍 Alamat: {selectedJob?.address}</Text>

              <TouchableOpacity 
                style={styles.btnWa} 
                onPress={() => Linking.openURL(`https://wa.me/${selectedJob?.customerPhone}`)}>
                <Text style={styles.btnText}>💬 Chat / Telpon Pembeli via WA</Text>
              </TouchableOpacity>
            </View>

            <View style={styles.card}>
              <Text style={styles.cardStore}>📷 UPLOAD NOTA FISIK WARUNG (MINIO)</Text>
              
              {receiptImage ? (
                <Image source={{ uri: receiptImage }} style={styles.previewImage} />
              ) : (
                <TouchableOpacity style={styles.btnPhoto} onPress={handlePickReceiptImage}>
                  <Text style={styles.btnText}>📸 AMBIL FOTO NOTA DARI KAMERA HP</Text>
                </TouchableOpacity>
              )}

              <Text style={{ color: '#94a3b8', fontSize: 12, marginTop: 12 }}>Total Harga Asli Belanja di Warung (Rp):</Text>
              <TextInput 
                style={styles.inputNumber} 
                keyboardType="numeric" 
                value={actualPrice} 
                onChangeText={setActualPrice}
              />

              <TouchableOpacity style={styles.btnSubmit} onPress={handleSubmitReceipt}>
                <Text style={styles.btnText}>🚀 SIMPAN NOTA & MULAI ANTAR</Text>
              </TouchableOpacity>
            </View>
          </View>
        )}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#0f172a' },
  centerContainer: { flex: 1, backgroundColor: '#0f172a', justifyContent: 'center', alignItems: 'center', padding: 20 },
  mainTitle: { fontSize: 24, font: '800', color: '#10b981', marginBottom: 6 },
  subTitle: { fontSize: 13, color: '#94a3b8', marginBottom: 24 },
  input: { width: '100%', backgroundColor: '#1e293b', color: '#fff', borderRadius: 12, padding: 14, marginBottom: 14 },
  inputNumber: { width: '100%', backgroundColor: '#0f172a', color: '#34d399', fontSize: 18, borderRadius: 10, padding: 10, marginTop: 4 },
  btnPrimary: { width: '100%', backgroundColor: '#059669', borderRadius: 12, padding: 16, alignItems: 'center' },
  header: { backgroundColor: '#047857', padding: 20, paddingTop: 45 },
  headerTitle: { fontSize: 18, color: '#fff' },
  headerStatus: { fontSize: 11, color: '#a7f3d0', marginTop: 2 },
  body: { padding: 16 },
  sectionTitle: { fontSize: 14, color: '#94a3b8', marginBottom: 12 },
  card: { backgroundColor: '#1e293b', borderRadius: 16, padding: 16, marginBottom: 16, borderWidth: 1, borderColor: 'rgba(255,255,255,0.05)' },
  cardCode: { fontSize: 12, color: '#10b981' },
  cardStore: { fontSize: 16, color: '#fff', marginVertical: 4 },
  cardText: { fontSize: 13, color: '#cbd5e1', marginTop: 2 },
  cardFee: { fontSize: 13, color: '#34d399', marginTop: 8 },
  btnClaim: { backgroundColor: '#059669', borderRadius: 10, padding: 12, alignItems: 'center', marginTop: 12 },
  btnWa: { backgroundColor: '#25D366', borderRadius: 10, padding: 10, alignItems: 'center', marginTop: 10 },
  btnPhoto: { backgroundColor: '#334155', borderStyle: 'dashed', borderWidth: 1, borderColor: '#94a3b8', borderRadius: 12, padding: 24, alignItems: 'center', marginVertical: 10 },
  previewImage: { width: '100%', height: 180, borderRadius: 12, marginVertical: 10 },
  btnSubmit: { backgroundColor: '#10b981', borderRadius: 12, padding: 14, alignItems: 'center', marginTop: 14 },
  btnText: { color: '#fff', fontSize: 13 }
});
