package com.example.fclient;

import androidx.activity.result.ActivityResult;
import androidx.activity.result.ActivityResultCallback;
import androidx.activity.result.ActivityResultLauncher;
import androidx.activity.result.ActivityResultRegistry;
import androidx.activity.result.contract.ActivityResultContracts;
import androidx.appcompat.app.AppCompatActivity;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.widget.TextView;
import android.widget.Toast;
import android.view.View;

import org.apache.commons.codec.DecoderException;
import org.apache.commons.codec.binary.Hex;

import com.example.fclient.databinding.ActivityMainBinding;

import java.nio.charset.StandardCharsets;

interface TransactionEvents {
    String enterPin(int ptc, String amount);
    void transactionResult(boolean result);
}

public class MainActivity extends AppCompatActivity implements TransactionEvents {
    ActivityResultLauncher activityResultLauncher;
    private String pin;
    @Override
    public String enterPin(int ptc, String amount) {
        pin = new String();
        Intent it = new Intent(MainActivity.this, PinpadActivity.class);
        it.putExtra("ptc", ptc);
        it.putExtra("amount", amount);
        synchronized (MainActivity.this) {
            activityResultLauncher.launch(it);
            try {
                MainActivity.this.wait();
            } catch (Exception ex) {
                //todo: log error
            }
        }
        return pin;
    }
    @Override
    public void transactionResult(boolean result) {
        runOnUiThread(()-> {
            Toast.makeText(MainActivity.this, result ? "ok" : "failed", Toast.LENGTH_SHORT).show();
        });
    }

    // Used to load the 'fclient' library on application startup.
    static {
        System.loadLibrary("fclient");
        System.loadLibrary("mbedcrypto");
        logForLibraries();

    }
    public void onButtonClick(View v)
    {
        byte[] trd = stringToHex("9F0206000000000100");
        transaction(trd);
    }

    public static byte[] stringToHex(String s)
    {
        byte[] hex;
        try
        {
            hex = Hex.decodeHex(s.toCharArray());
        }
        catch (DecoderException ex)
        {
            hex = null;
        }
        return hex;
    }

    private ActivityMainBinding binding;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        binding = ActivityMainBinding.inflate(getLayoutInflater());
        setContentView(binding.getRoot());

        int res = initRng();
        byte[] key = randomBytes(16);

        /*
        String test = stringFromJNI();




        byte[] original = "hello pasha".getBytes();
        byte[] encrypted = encrypt(key, original);
        byte[] decrypted = decrypt(key, encrypted);

        // Логируем первые 4 байта для примера
        logForInt(original.length);
        logForInt(encrypted.length);           // длина зашифрованных данных
        logForInt(decrypted.length);

        logByteArray(original,   "оригинал");
        logByteArray(encrypted,  "зашифровано");
        logByteArray(decrypted,  "расшифровано");
        Toast.makeText(this,
                "Расшифровано:\n" + new String(decrypted, StandardCharsets.UTF_8),
                Toast.LENGTH_LONG).show();

        // Example of a call to a native method
        //TextView tv = binding.sampleText;
        //tv.setText(stringFromJNI());
        /*
    */
        activityResultLauncher  = registerForActivityResult(
                new ActivityResultContracts.StartActivityForResult(),
                new ActivityResultCallback<ActivityResult>() {
                    @Override
                    public void onActivityResult(ActivityResult result) {
                        if (result.getResultCode() == Activity.RESULT_OK) {
                            Intent data = result.getData();
                            // обработка результата
                            //String pin = data.getStringExtra("pin");
                            //Toast.makeText(MainActivity.this, pin, Toast.LENGTH_SHORT).show();
                            pin = data.getStringExtra("pin");
                            synchronized (MainActivity.this) {
                                MainActivity.this.notifyAll();
                            }
                        }
                    }
                });

    }


    /*
     * A native method that is implemented by the 'fclient' native library,
     * which is packaged with this application.
     */
    //нативные функции
    public native String stringFromJNI(); // получение строки из C++
    public static native int initRng();   // инициализация генератора случайных чисел
    public static native byte[] randomBytes(int no); // получение массива случайных байт
    public static native void logForLibraries(); //лог о том что библиотеки хорошо работают
    public static native  void logForInt(int num); //лог вывода байта
    public static native byte[] encrypt(byte[] key, byte[] data); // шифрование
    public static native byte[] decrypt(byte[] key, byte[] data); // дешифровка
    public static native void logByteArray(byte[] array, String label);

    public native boolean transaction(byte[] trd);
}

