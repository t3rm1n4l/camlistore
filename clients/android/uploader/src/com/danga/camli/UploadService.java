package com.danga.camli;

import java.io.FileNotFoundException;
import java.util.HashSet;
import java.util.LinkedList;
import java.util.Set;

import android.app.Service;
import android.content.ContentResolver;
import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.net.Uri;
import android.os.IBinder;
import android.os.ParcelFileDescriptor;
import android.os.PowerManager;
import android.os.RemoteException;
import android.util.Log;

public class UploadService extends Service {
    private static final String TAG = "UploadService";

    // Guarded by 'this':
    private boolean mUploading = false;
    private UploadThread mUploadThread = null;
    final Set<QueuedFile> mQueueSet = new HashSet<QueuedFile>();
    final LinkedList<QueuedFile> mQueueList = new LinkedList<QueuedFile>();

	@Override
    public IBinder onBind(Intent intent) {
        return service;
	}

    // Called by UploadThread to get stuff to do. Caller owns the returned new
    // LinkedList. Doesn't return null.
    LinkedList<QueuedFile> uploadQueue() {
        synchronized (this) {
            LinkedList<QueuedFile> copy = new LinkedList<QueuedFile>();
            copy.addAll(mQueueList);
            return copy;
        }
    }

    void onUploadThreadEnding() {
        synchronized (this) {
            mUploadThread = null;
            mUploading = false;
        }
    }

    void onUploadComplete(QueuedFile qf) {
        synchronized (this) {
            boolean removedSet = mQueueSet.remove(qf);
            boolean removedList = mQueueList.remove(qf); // TODO: ghetto, linear
                                                         // scan
            Log.d(TAG, "removing of " + qf + "; removedSet=" + removedSet + "; removedList="
                    + removedList);
        }
    }

    ParcelFileDescriptor getFileDescriptor(Uri uri) {
        ContentResolver cr = getContentResolver();
        try {
            return cr.openFileDescriptor(uri, "r");
        } catch (FileNotFoundException e) {
            Log.w(TAG, "FileNotFound for " + uri, e);
            return null;
        }
    }

    private final IUploadService.Stub service = new IUploadService.Stub() {

        public boolean enqueueUpload(Uri uri) throws RemoteException {
            SharedPreferences sp = getSharedPreferences(Preferences.NAME, 0);
            HostPort hp = new HostPort(sp.getString(Preferences.HOST, ""));
            if (!hp.isValid()) {
                return false;
            }

            ParcelFileDescriptor pfd = getFileDescriptor(uri);

            String sha1 = Util.getSha1(pfd.getFileDescriptor());
            Log.d(TAG, "sha1 of file is: " + sha1);
            Log.d(TAG, "size of file is: " + pfd.getStatSize());
            QueuedFile qf = new QueuedFile(sha1, uri);

            synchronized (UploadService.this) {
                if (mQueueSet.contains(qf)) {
                    return false;
                }
                mQueueSet.add(qf);
                mQueueList.add(qf);
                if (!mUploading) {
                    resume();
                }
            }
            return true;
        }

        public boolean isUploading() throws RemoteException {
            synchronized (UploadService.this) {
                return mUploading;
            }
        }

        public void registerCallback(IStatusCallback cb) throws RemoteException {
            // TODO Auto-generated method stub

        }

        public boolean resume() throws RemoteException {
            synchronized (UploadService.this) {
                if (mUploadThread != null) {
                    return false;
                }
                mUploading = true;

                SharedPreferences sp = getSharedPreferences(Preferences.NAME, 0);
                HostPort hp = new HostPort(sp.getString(Preferences.HOST, ""));
                if (!hp.isValid()) {
                    return false;
                }
                String password = sp.getString(Preferences.PASSWORD, "");

                PowerManager pm = (PowerManager) getSystemService(Context.POWER_SERVICE);
                final PowerManager.WakeLock wl = pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "Camli Upload");
                wl.acquire();
                mUploadThread = new UploadThread(UploadService.this, hp,
                        password);

                // Start a thread to release the wakelock...
                final Thread threadToWatch = mUploadThread;
                new Thread() {
                    @Override public void run() {
                        try {
                            threadToWatch.join();
                        } catch (InterruptedException e) {
                        }
                        Log.d(TAG, "UploadThread done; releasing the wakelock");
                        wl.release();
                    }
                }.start();
                mUploadThread.start();
                return true;
            }
        }

        public boolean pause() throws RemoteException {
            synchronized (UploadService.this) {
                if (mUploadThread != null) {
                    mUploadThread.stopPlease();
                    return true;
                }
                return false;
            }
        }

        public void unregisterCallback(IStatusCallback cb)
                throws RemoteException {
            // TODO Auto-generated method stub

        }

        public int queueSize() throws RemoteException {
            synchronized (UploadService.this) {
                return mQueueList.size();
            }
        }
    };
}
